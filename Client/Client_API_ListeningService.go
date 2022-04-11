package client

import (
	pc "assignment1/main/protocchubby"
	"context"
	"fmt"
)

// SendEventMessage responds to an event message
func (c *Client) SendEventMessage(ctx context.Context, eventMsg *pc.EventMessage) (*pc.EventMessage, error) {
	fmt.Printf("> Client %s received an event notification \n", eventMsg.Type.String())
	c.ClientCacheValidation[eventMsg.FileName] = false
	return &pc.EventMessage{Type: pc.EventMessage_Ack}, nil
}

// SendClientMessage is a Channel for ClientMessages
func (c *Client) SendClientMessage(ctx context.Context, CliMsg *pc.ClientMessage) (*pc.ClientMessage, error) {
	switch CliMsg.Type {
	default:
		fmt.Printf("> Unimplemented.\n")
		return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Error}, nil
	}

	return &pc.ClientMessage{ClientID: CliMsg.ClientID, Type: pc.ClientMessage_Ack}, nil
}
