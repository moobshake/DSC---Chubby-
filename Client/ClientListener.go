package client

import (
	"context"
	"fmt"
	"log"
	"net"

	NC "assignment1/main/NodeComm"

	"google.golang.org/grpc"
)

//startClientListener starts a *grpc.serve server as a listener.
func (c *Client) startClientListener() {
	fullAddress := c.ClientAdd.IP + ":" + c.ClientAdd.Port
	fmt.Println("Starting listener at ", fullAddress)
	lis, err := net.Listen("tcp", fullAddress)
	if err != nil {
		log.Fatalf("Failed to hook into: %s. %v", fullAddress, err)
	}

	s := Client{}
	gServer := grpc.NewServer()
	NC.RegisterNodeCommServiceServer(gServer, &s)

	if err := gServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}

// Respond to an event message
func (c *Client) SendEventMessage(ctx context.Context, eventMsg *NC.EventMessage) (*NC.EventMessage, error) {
	fmt.Printf("> Client %s receieved an event notification \n", eventMsg.Type.String())
	return &NC.EventMessage{Type: NC.EventMessage_Ack}, nil
}
