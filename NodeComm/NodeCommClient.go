package nodecomm

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//<><><><><><><><><><><><><><><><><><><><><>
//<><><> CLient side methods corresponding to the RPC Methods <><><>

//<> "Internal" Control API <>

//DispatchShutdown sends a shutdown signal to the server
func (n *Node) DispatchShutdown() bool {
	conn, err := connectTo(n.myPRecord.Address, n.myPRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)

	cMsg := ControlMessage{Type: int32(StopListening)}

	response, err := c.Shutdown(context.Background(), &cMsg)
	if err != nil {
		fmt.Println("Error dispatching shutdown signal:", err)
	}
	if response.Type == int32(Okay) {
		return true
	}
	return false
}

//DispatchControlMessage dispatches a control message to the server
func (n *Node) DispatchControlMessage(cMsg *ControlMessage) *ControlMessage {
	conn, err := connectTo(n.myPRecord.Address, n.myPRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)
	response, err := c.SendControlMessage(context.Background(), cMsg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}
	return response
}

//<> "External" Peer-to-Peer API <>

//DO NOT PUT THE BAD NODE HANDLER HERE TO AVOID A NEVER ENDING LOOP
//THIS SHOULD REALLY BE RENAMED TO CHECKALIVE
//DispatchKeepAlive sends a NodeMessage as a keepalive message to the node of the given PeerRecord.
func (n *Node) DispatchKeepAlve(destPRec *PeerRecord) bool {
	if destPRec == nil {
		return false
	}
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error connecting:", err)
		}

		return false
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)

	gibberish := strconv.Itoa(rand.Int())

	nodeMessage := NodeMessage{FromPRecord: n.myPRecord, Type: int32(KeepAlive), Comment: gibberish}

	response, err := c.KeepAlive(context.Background(), &nodeMessage)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error dispatching keepalive:", err)
		}

		return false
	}
	if response.Type == int32(KeepAlive) && response.Comment == gibberish {
		return true
	}
	return false
}

//DispatchMessage dispatches the given NodeMessage to the node of the given PeerRecord
func (n *Node) DispatchMessage(destPRec *PeerRecord, nodeMessage *NodeMessage) *NodeMessage {
	if destPRec == nil {
		return nil
	}
	nodeMessage.FromPRecord = n.myPRecord
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error connecting:", err)
		}

		n.badNodeHandler(destPRec)
		return nil
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)

	response, err := c.SendMessage(context.Background(), nodeMessage)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error dispatching messsage:", err)
		}

		n.badNodeHandler(destPRec)
		return nil
	}
	return response
}

//DispatchCoordinationMessage dispatches a given CoordinationMessage to the node of the given PeerRecord.
func (n *Node) DispatchCoordinationMessage(destPRec *PeerRecord, nCoMsg *CoordinationMessage) *CoordinationMessage {
	if destPRec == nil {
		return nil
	}
	nCoMsg.FromPRecord = n.myPRecord
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error connecting:", err)
		}
		n.badNodeHandler(destPRec)
		return nil
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)

	response, err := c.SendCoordinationMessage(context.Background(), nCoMsg)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error dispatching coordination messsage:", err)
		}

		n.badNodeHandler(destPRec)
		return nil
	}

	if response.Type == int32(NotMaster) && n.IsMaster() { //Network is fractured. Fix it.
		n.startElection()
	}

	return response
}

//Convenience methods

func (n *Node) BroadcastCoordinationMessage(nCoMsg *CoordinationMessage) {
	for _, pRec := range n.peerRecords {
		n.DispatchCoordinationMessage(pRec, nCoMsg)
	}
}

func (n *Node) BroadcastElectionResults() {
	nBCoMsg := CoordinationMessage{PeerRecords: append(n.peerRecords, n.myPRecord), Type: int32(ElectionResult)}
	n.BroadcastCoordinationMessage(&nBCoMsg)
}

func (n *Node) BroadcastPeerInformation() {
	nBCoMsg := CoordinationMessage{PeerRecords: append(n.peerRecords, n.myPRecord), Type: int32(PeerInformation)}
	n.BroadcastCoordinationMessage(&nBCoMsg)
}

//Utility
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}
