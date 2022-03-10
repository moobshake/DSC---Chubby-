package main

import (
	NC "assignment1/main/NodeComm"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//Put methods that directly interface with the server RPC methods here
//Ideally, they start with Dispatch...

//Refer to NodCommFlags.go to see what flags are applicable for NC.ClientMessage

func (c *Client) DispatchClientMessage(destPRec *NC.PeerRecord, CliMsg *NC.ClientMessage) *NC.ClientMessage {
	//Implement
	if destPRec == nil {
		return nil
	}
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		return nil
	}
	defer conn.Close()

	cli := NC.NewNodeCommServiceClient(conn)

	response, err := cli.SentClientMessage(context.Background(), CliMsg)
	if err != nil {
		return nil
	}
	return response
}

//Utility
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}
