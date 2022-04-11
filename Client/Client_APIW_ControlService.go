package client

import (
	"context"
	"fmt"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

//DispatchShutdown sends a shutdown signal to the server
func (c *Client) DispatchShutdown() bool {
	conn, err := connectTo(c.ClientAdd.IP, c.ClientAdd.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cClient := pc.NewClientControlServiceClient(conn)

	cMsg := pc.ControlMessage{Type: pc.ControlMessage_StopListening}

	response, err := cClient.Shutdown(context.Background(), &cMsg)
	if err != nil {
		fmt.Println("Error dispatching shutdown signal:", err)
	}
	if response.Type == pc.ControlMessage_Okay {
		return true
	}
	return false
}

// Send a ClientMessage to the client's own listener
func (c *Client) DispatchControlClientMessage(CliMsg *pc.ClientMessage) *pc.ClientMessage {
	conn, err := connectTo(c.ClientAdd.IP, c.ClientAdd.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cConn := pc.NewClientControlServiceClient(conn)
	response, err := cConn.SendControlMessage(context.Background(), CliMsg)
	if err != nil {
		fmt.Println("Error dispatching client control message:", err)
	}
	return response
}
