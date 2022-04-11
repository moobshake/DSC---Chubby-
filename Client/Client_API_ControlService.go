package client

import (
	pc "assignment1/main/protocchubby"
	"context"
	"fmt"
)

//Shutdown - Unimplemented as I found that a graceful shutdown is unimportant for this.
func (c *Client) Shutdown(ctx context.Context, cMsg *pc.ControlMessage) (*pc.ControlMessage, error) {
	if cMsg.Type != pc.ControlMessage_StopListening {
		return &pc.ControlMessage{Type: pc.ControlMessage_Error, Comment: "Wrong channel."}, nil
	}
	return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
}

// SendClientMessage is a Channel for ClientMessages
func (c *Client) SendControlMessage(ctx context.Context, CliMsg *pc.ClientMessage) (*pc.ClientMessage, error) {
	switch CliMsg.Type {
	case pc.ClientMessage_Init:
		go c.LockChecker()
	case pc.ClientMessage_FindMaster:
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)
		c.FindMaster()
	case pc.ClientMessage_FileWrite:
		c.DispatchClientWriteRequest(CliMsg.StringMessages /*modifyFile*/, true)
	case pc.ClientMessage_FileRead:
		c.ClientReadRequest(CliMsg.StringMessages)

	case pc.ClientMessage_WriteLock:
		c.ClientRequest(REQ_LOCK, WRITE_CLI, CliMsg.StringMessages)

	case pc.ClientMessage_ReadLock:
		c.ClientRequest(REQ_LOCK, READ_CLI, CliMsg.StringMessages)

	case pc.ClientMessage_ReleaseLock:
		c.ClientRequest(REL_LOCK, CliMsg.StringMessages)

	case pc.ClientMessage_ListFile:
		c.ClientRequest(LIST_FILE_CLI)
	case pc.ClientMessage_ListLocks:
		c.ListLocks()
	default:
		fmt.Printf("> Client CLI requesting for something that is not available %s\n", CliMsg.Type.String())
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Error}, nil
	}

	return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Ack}, nil
}
