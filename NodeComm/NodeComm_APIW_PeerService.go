package nodecomm

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

//DispatchKeepAlive sends a pc.NodeMessage as a keepalive message to the node of the given pc.PeerRecord.
//DO NOT PUT THE BAD NODE HANDLER HERE TO AVOID A NEVER ENDING LOOP
//THIS SHOULD REALLY BE RENAMED TO CHECKALIVE
func (n *Node) DispatchKeepAlive(destPRec *pc.PeerRecord) bool {
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

	c := pc.NewNodeCommPeerServiceClient(conn)

	gibberish := strconv.Itoa(rand.Int())

	nodeMessage := pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_KeepAlive, Comment: gibberish}

	response, err := c.KeepAlive(context.Background(), &nodeMessage)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error dispatching keepalive:", err)
		}

		return false
	}
	if response.Type == pc.NodeMessage_KeepAlive && response.Comment == gibberish {
		return true
	}
	return false
}

//DispatchMessage dispatches the given pc.NodeMessage to the node of the given pc.PeerRecord
func (n *Node) DispatchMessage(destPRec *pc.PeerRecord, nodeMessage *pc.NodeMessage) *pc.NodeMessage {
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

	c := pc.NewNodeCommPeerServiceClient(conn)

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

//DispatchCoordinationMessage dispatches a given CoordinationMessage to the node of the given pc.PeerRecord.
func (n *Node) DispatchCoordinationMessage(destPRec *pc.PeerRecord, nCoMsg *pc.CoordinationMessage) *pc.CoordinationMessage {
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

	c := pc.NewNodeCommPeerServiceClient(conn)

	response, err := c.SendCoordinationMessage(context.Background(), nCoMsg)
	if err != nil {
		if n.verbose == 2 {
			fmt.Println("Error dispatching coordination messsage:", err)
		}

		n.badNodeHandler(destPRec)
		return nil
	}

	if response.Type == pc.CoordinationMessage_NotMaster && n.IsMaster() { //Network is fractured. Fix it.
		n.startElection()
	}

	return response
}

// Sends a Server Message to a replica
func (n *Node) DispatchServerMessage(destPRec *pc.PeerRecord, readCheckMsg *pc.ServerMessage) *pc.ServerMessage {
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return nil
	}
	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	replicaMsg, err := cli.EstablishReplicaConsensus(context.Background(), readCheckMsg)

	if err != nil {
		fmt.Println("DispatchServerMessage: ERROR", err)
		return nil
	}

	return replicaMsg
}

//Convenience Dispatch methods - Should really be in another file

//BroadcastCoordinationMessage calls DispatchCoordinationMessage for all known peers
func (n *Node) BroadcastCoordinationMessage(nCoMsg *pc.CoordinationMessage) {
	for _, pRec := range n.peerRecords {
		n.DispatchCoordinationMessage(pRec, nCoMsg)
	}
}

//BroadcastElectionResults sends elections results to all known peers
func (n *Node) BroadcastElectionResults() {
	nBCoMsg := pc.CoordinationMessage{PeerRecords: append(n.peerRecords, n.myPRecord), Type: pc.CoordinationMessage_ElectionResult}
	n.BroadcastCoordinationMessage(&nBCoMsg)
}

//BroadcastPeerInformation sends peer information to all known peers
func (n *Node) BroadcastPeerInformation() {
	nBCoMsg := pc.CoordinationMessage{PeerRecords: append(n.peerRecords, n.myPRecord), Type: pc.CoordinationMessage_PeerInformation}
	n.BroadcastCoordinationMessage(&nBCoMsg)
}
