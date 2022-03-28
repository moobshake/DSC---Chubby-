package client

import (
	pc "assignment1/main/protocchubby"
	"context"
)

//Shutdown - Unimplemented as I found that a graceful shutdown is unimportant for this.
func (c *Client) Shutdown(ctx context.Context, cMsg *pc.ControlMessage) (*pc.ControlMessage, error) {
	if cMsg.Type != pc.ControlMessage_StopListening {
		return &pc.ControlMessage{Type: pc.ControlMessage_Error, Comment: "Wrong channel."}, nil
	}
	return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
}
