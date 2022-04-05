package nodecomm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pc "assignment1/main/protocchubby"
)

//<><><> RPC Send Methods - These methods RECEIVE messages <><><>

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

	// TODO: Ask YH to change from stringmessages to the lock message
	case pc.ClientMessage_WriteLock:
		isAvail, seq, timestamp, lockdelay := n.AcquireWriteLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_WriteLock, StringMessages: nodeReply, Lock: &pc.LockMessage{Type: pc.LockMessage_WriteLock, Sequencer: nodeReply, TimeStamp: timestamp, LockDelay: int32(lockdelay)}}, nil

	// TODO: Ask YH to change from stringmessages to the lock message1
	case pc.ClientMessage_ReadLock:
		isAvail, seq, timestamp, lockdelay := n.AcquireReadLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_ReadLock, StringMessages: nodeReply, Lock: &pc.LockMessage{Type: pc.LockMessage_WriteLock, Sequencer: nodeReply, TimeStamp: timestamp, LockDelay: int32(lockdelay)}}, nil
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
		// Get the file from the local dir in batches
		localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
		file, err := os.Open(localFilePath)
		if err != nil {
			fmt.Println("CLIENT FILE READ REQUEST ERROR:", err)
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

// Receive a stream of write messages from the client.
func (n *Node) SendWriteRequest(stream pc.NodeCommListeningService_SendWriteRequestServer) error {
	var writeRequestMessage *pc.ClientMessage

	writeRequestMessage, err := stream.Recv()
	if err != nil {
		return err
	}
	if writeRequestMessage.Type == pc.ClientMessage_FileWrite {
		return n.handleClientWriteRequest(stream, writeRequestMessage)
	} else if writeRequestMessage.Type == pc.ClientMessage_ReplicaWrites {
		return n.handleMasterToReplicatWriteRequest(stream, writeRequestMessage)
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
