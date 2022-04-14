package nodecomm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	pc "assignment1/main/protocchubby"
)

//<><><> RPC Send Methods - These methods RECEIVE messages <><><>

//KeepAlive - Used to check if a node is alive or not.
func (n *Node) KeepAliveForClient(ctx context.Context, inMsg *pc.ClientMessage) (*pc.ClientMessage, error) {
	switch inMsg.Type {
	case pc.ClientMessage_KeepAlive:
		nMsg := pc.ClientMessage{Type: pc.ClientMessage_Ack}
		n.NoteClientIsAlive(int(inMsg.ClientID))
		return &nMsg, nil
	default:
		//Received a non-keepalive message on the keepalive channel
		return &pc.ClientMessage{Type: pc.ClientMessage_Empty}, nil
	}
}

// SendClientMessage is a Channel for ClientMessages
// Note: Read Requests are not processed here
func (n *Node) SendClientMessage(ctx context.Context, CliMsg *pc.ClientMessage) (*pc.ClientMessage, error) {

	var nodeReply string

	// If node is not the master, do not handle any of the messages
	// just tell the client who is
	if !n.IsMaster() {
		return n.getRedirectionCliMsg(CliMsg.ClientID), nil
	}

	switch CliMsg.Type {
	case pc.ClientMessage_FindMaster:
		// Find master
		// Replies with master's address
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)
		return &pc.ClientMessage{
			ClientID:      CliMsg.ClientID,
			Type:          pc.ClientMessage_ConfirmCoordinator,
			Spare:         int32(n.idOfMaster),
			ClientAddress: n.getPeerRecord(n.idOfMaster, true),
		}, nil

	case pc.ClientMessage_SubscribeFileModification, pc.ClientMessage_SubscribeLockAquisition,
		pc.ClientMessage_SubscribeLockConflict, pc.ClientMessage_SubscribeMasterFailover:
		fmt.Printf("> Client %d requesting to subscibe: %s\n", CliMsg.ClientID, CliMsg.Type.String())
		n.ClientSubscriptionsHandler(CliMsg)

	case pc.ClientMessage_ListFile:
		fmt.Printf("> Client %d requesting to list files\n", CliMsg.ClientID)
		f := list_files(n.nodeDataPath)
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Ack, StringMessages: f}, nil

	case pc.ClientMessage_WriteLock:
		isAvail, seq, timestamp, lockdelay := n.AcquireWriteLock(CliMsg.StringMessages, int(CliMsg.ClientID), int(CliMsg.Spare))
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}
		// Send locks to replicas
		l := &pc.LockMessage{Type: pc.LockMessage_WriteLock, Sequencer: nodeReply, TimeStamp: timestamp, LockDelay: int32(lockdelay)}

		if isAvail && !n.SendRequestToReplicas(&pc.ServerMessage{Type: pc.ServerMessage_ReqLock, Lock: l, StringMessages: strconv.Itoa(int(CliMsg.ClientID))}) {
			fmt.Printf("Forwarding write lock request from %d to replicas\n", CliMsg.ClientID)
			return &pc.ClientMessage{Type: pc.ClientMessage_Error, StringMessages: "Majority of replicas do not agree on the read file."}, nil
		}

		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_WriteLock, StringMessages: nodeReply, Lock: l}, nil

	case pc.ClientMessage_ReadLock:
		isAvail, seq, timestamp, lockdelay := n.AcquireReadLock(CliMsg.StringMessages, int(CliMsg.ClientID), int(CliMsg.Spare))
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}

		l := &pc.LockMessage{Type: pc.LockMessage_ReadLock, Sequencer: nodeReply, TimeStamp: timestamp, LockDelay: int32(lockdelay)}

		// Send locks to replicas
		fmt.Printf("Forwarding read lock request from %d to replicas\n", CliMsg.ClientID)
		if isAvail && !n.SendRequestToReplicas(&pc.ServerMessage{Type: pc.ServerMessage_ReqLock, Lock: l, StringMessages: strconv.Itoa(int(CliMsg.ClientID))}) {
			return &pc.ClientMessage{Type: pc.ClientMessage_Error, StringMessages: "Majority of replicas do not agree on the read file."}, nil
		}

		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_ReadLock, StringMessages: nodeReply, Lock: l}, nil

	case pc.ClientMessage_ReleaseLock:
		// check if lock exist and should be released
		message := n.ReleaseLockChecker(int(CliMsg.ClientID), CliMsg.Lock.Type, CliMsg.Lock.Sequencer)
		if !n.SendRequestToReplicas(&pc.ServerMessage{Type: pc.ServerMessage_RelLock, Lock: CliMsg.Lock, StringMessages: strconv.Itoa(int(CliMsg.ClientID))}) {
			return &pc.ClientMessage{Type: pc.ClientMessage_Error, StringMessages: "Majority of replicas do not agree to release lock"}, nil
		}

		return &pc.ClientMessage{Type: pc.ClientMessage_ReleaseLock, StringMessages: message}, nil

	default:
		fmt.Printf("> Client %d requesting for something that is not available %s\n", CliMsg.ClientID, CliMsg.Type.String())
	}

	return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Error, StringMessages: nodeReply}, nil
}

//SendEventMessage is a ...
func (n *Node) SendEventMessage(ctx context.Context, eMsg *pc.EventMessage) (*pc.EventMessage, error) {
	return nil, nil
}

// SendReadRequest streams the required file from local data to the client in batches
func (n *Node) SendReadRequest(CliMsg *pc.ClientMessage, stream pc.NodeCommListeningService_SendReadRequestServer) error {
	fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)

	if !n.IsMaster() {
		// Return a master redirection message
		cliMsg := n.getRedirectionCliMsg(CliMsg.ClientID)
		if err := stream.Send(cliMsg); err != nil {
			return err
		}

		return nil
	}

	if n.validateReadRequest(CliMsg) {
		// Check if the file is consistent across the majority of replicas
		// Create the serverMsg with the file name to check.
		replicaMsg := pc.ServerMessage{
			Type:           pc.ServerMessage_ReplicaReadCheck,
			StringMessages: CliMsg.StringMessages,
		}
		if !n.SendRequestToReplicas(&replicaMsg) {
			// Return an Error
			cliMsg := pc.ClientMessage{
				Type:           pc.ClientMessage_Error,
				StringMessages: "Majority of replicas do not agree on the read file.",
			}
			if err := stream.Send(&cliMsg); err != nil {
				return err
			}
			return nil
		}

		// Get the file from the local dir in batches
		localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
		file, err := os.Open(localFilePath)
		if err != nil {
			fmt.Println("CLIENT FILE READ REQUEST ERROR:", err)
			return nil
		}
		defer file.Close()

		buffer := make([]byte, READ_MAX_BYTE_SIZE)

		for {
			numBytes, err := file.Read(buffer)

			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
			fileContent := pc.FileBodyMessage{
				Type:        pc.FileBodyMessage_ReadMode,
				FileName:    CliMsg.StringMessages,
				FileContent: buffer[:numBytes],
			}

			cliMsg := pc.ClientMessage{
				Type:     pc.ClientMessage_FileRead,
				FileBody: &fileContent,
			}
			if err := stream.Send(&cliMsg); err != nil {
				return err
			}
		}

		return nil
	} else {
		// Return an invalid lock error
		cliMsg := pc.ClientMessage{
			Type: pc.ClientMessage_InvalidLock,
		}
		if err := stream.Send(&cliMsg); err != nil {
			return err
		}
		return nil
	}
}

// Receive a stream of write messages from the client or from the master for replication.
func (n *Node) SendWriteRequest(stream pc.NodeCommListeningService_SendWriteRequestServer) error {
	var writeRequestMessage *pc.ClientMessage

	writeRequestMessage, err := stream.Recv()
	if err != nil {
		return err
	}
	if writeRequestMessage.Type == pc.ClientMessage_FileWrite {
		return n.handleClientWriteRequest(stream, writeRequestMessage)
	}

	return nil
}

func (n *Node) getRedirectionCliMsg(clientId int32) *pc.ClientMessage {
	return &pc.ClientMessage{
		ClientID:      clientId,
		Type:          pc.ClientMessage_RedirectToCoordinator,
		Spare:         int32(n.idOfMaster),
		ClientAddress: n.getPeerRecord(n.idOfMaster, true),
	}
}
