package client

import (
	NC "assignment1/main/NodeComm"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//Put methods that directly interface with the server RPC methods here
//Ideally, they start with Dispatch...

//Refer to NodCommFlags.go to see what flags are applicable for NC.ClientMessage

func (c *Client) DispatchClientMessage(destPRec *NC.PeerRecord, CliMsg *NC.ClientMessage) *NC.ServerMessage {
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

	response, err := cli.SendClientMessage(context.Background(), CliMsg)
	if err != nil {
		return nil
	}
	return response
}

// Mock Client Listener based off Jiawei's
// func (c *Client) clientListener() {

// 	if c.ClientID == -1 {
// 		fmt.Println("Unable to start without ClientId set!")
// 		return
// 	}

// 	cAddr := fmt.Sprintf("%s:%s", c.ClientAdd.IP, c.ClientAdd.Port)
// 	fmt.Printf("Client %d started listening at %s\n", c.ClientID, cAddr)

// 	lis, err := net.Listen("tcp", cAddr)
// 	if err != nil {
// 		log.Fatalf("Failed to hook into %s. %v", cAddr, err)
// 	}

// 	// Ask Jia Wei if I should be registering this to NodeCommServiceServer

// }

//Utility
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}
