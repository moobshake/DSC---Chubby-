package nodecomm

import (
	"context"
	"fmt"

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

	if CliMsg.Type == pc.ClientMessage_FileRead {
		return n.handleReadRequestFromClient(CliMsg, stream)
	} else if CliMsg.Type == pc.ClientMessage_ReplicaReadCheck {
		return n.handleReadRequestFromMaster(CliMsg, stream)
	} else {
		// Return an error
		cliMsg := pc.ClientMessage{
			Type: pc.ClientMessage_Error,
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
