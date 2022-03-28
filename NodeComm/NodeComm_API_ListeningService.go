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

	//If this server is not the master
	if !(n.IsMaster()) {
		switch CliMsg.Type {
		case pc.ClientMessage_FindMaster:
			n.DispatchClientMessage(CliMsg.ClientAddress, &pc.ClientMessage{Type: pc.ClientMessage_RedirectToCoordinator, Spare: int32(n.idOfMaster), ClientAddress: n.getPeerRecord(n.idOfMaster, true)})
		default:
			n.DispatchClientMessage(CliMsg.ClientAddress, &pc.ClientMessage{Type: pc.ClientMessage_RedirectToCoordinator, Spare: int32(n.idOfMaster), ClientAddress: n.getPeerRecord(n.idOfMaster, true)})
		}
	}

	// Replies with master address
	switch CliMsg.Type {
	case pc.ClientMessage_FindMaster:
		// Find master
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)

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
		isAvail, seq := n.AcquireWriteLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_WriteLock, StringMessages: nodeReply}, nil

	// TODO: Ask YH to change from stringmessages to the lock message
	case pc.ClientMessage_ReadLock:
		isAvail, seq := n.AcquireReadLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_ReadLock, StringMessages: nodeReply}, nil
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

			if err := stream.Send(&fileContent); err != nil {
				return err
			}
		}

		return nil
	} else {
		// Return an invalid lock error
		fileContent := pc.FileBodyMessage{
			Type: pc.FileBodyMessage_InvalidLock,
		}
		if err := stream.Send(&fileContent); err != nil {
			return err
		}
		return nil
	}

}

// Receive a stream of write messages from the client.
func (n *Node) SendWriteRequest(stream pc.NodeCommListeningService_SendWriteRequestServer) error {
	var writeRequestMessage *pc.ClientMessage

	// This is the first message from the client that should
	// contain a valid write lock.
	writeRequestMessage, err := stream.Recv()
	if err != nil {
		return err
	}

	// Validate write lock
	// TODO(Hannah): change to appropriate function
	if n.validateWriteLock() {
		n.writeToLocalFile(writeRequestMessage, true)

		// Keep listening for more messages from the client in case
		// the file is very big.
		for {
			writeRequestMessage, err = stream.Recv()
			if err == io.EOF {
				return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_FileWrite})
			}
			if err != nil {
				return err
			}

			n.writeToLocalFile(writeRequestMessage, false)
		}
	} else {
		return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_InvalidLock})
	}
}
