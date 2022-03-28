package client

import (
	pc "assignment1/main/protocchubby"
	"context"
	"fmt"
)

// SendClientMessage is a Channel for ClientMessages
// Note: Read Requests are not processed here
func (c *Client) SendClientMessage(ctx context.Context, CliMsg *pc.ClientMessage) (*pc.ClientMessage, error) {
	return nil, nil
}

// SendEventMessage responds to an event message
func (c *Client) SendEventMessage(ctx context.Context, eventMsg *pc.EventMessage) (*pc.EventMessage, error) {
	fmt.Printf("> Client %s received an event notification \n", eventMsg.Type.String())
	return &pc.EventMessage{Type: pc.EventMessage_Ack}, nil
}
