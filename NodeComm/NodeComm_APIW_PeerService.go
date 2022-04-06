package nodecomm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"

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

//Dispatch methods for SendWriteForward

// Write requests should be acknowledged by the majority of replicas
// DispatchWriteForward sends a stream of messages from the master to the replicas.
// The replicas should save the file to their local copies and return an OK if they are done
// Otherwise, send back an error.
func (n *Node) DispatchWriteForward(writeMsgBuffer []*pc.ClientMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {
	// create stream for message sending
	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	stream, err := cli.SendWriteForward(context.Background())
	if err != nil {
		fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		return
	}

	for _, cliWriteMsg := range writeMsgBuffer {
		// Either the end of file or an error. Break.
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		// The file content to embed in the request
		cliWriteMsg.Type = pc.ClientMessage_ReplicaWrites

		if err := stream.Send(cliWriteMsg); err != nil {
			fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
	}

	if reply.Type == pc.ClientMessage_Ack {
		replyChan <- true
	} else {
		replyChan <- false
	}

	fmt.Println("Master write to replica reply from replica:", reply.Type)

}

// DispatchFileToReplica sends a stream of messages to a replica.
func (n *Node) DispatchFileToReplica(dPRec *pc.PeerRecord, filePath string) {
	// create stream for message sending
	conn, err := connectTo(dPRec.Address, dPRec.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	stream, err := cli.SendWriteForward(context.Background())

	// Retrive the file to send
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error sending file to replica.", err)
		return
	}

	defer file.Close()

	// Get the file from the cache dir in batches
	buffer := make([]byte, READ_MAX_BYTE_SIZE)

	// All messages the client sends will have both the lock
	// and the file to be modified.
	for {
		numBytes, err := file.Read(buffer)

		// Either the end of file or an error. Break.
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		tokenisedFilePath := strings.Split(filePath, "/")
		fileName := tokenisedFilePath[len(tokenisedFilePath)-1]

		// The file content to embed in the request
		fileContent := pc.FileBodyMessage{
			Type:        pc.FileBodyMessage_WriteMode,
			FileName:    fileName,
			FileContent: buffer[:numBytes],
		}

		cliMsg := pc.ClientMessage{
			ClientID: int32(n.myPRecord.Id),
			Type:     pc.ClientMessage_FileWrite,
			// The name of the file to write
			StringMessages: fileName,
			FileBody:       &fileContent,
			ClientAddress:  n.myPRecord,
		}

		if err := stream.Send(&cliMsg); err != nil {
			fmt.Println("Error when Master tried to send file to replica:", err)
		}
	}
	stream.CloseAndRecv()
}

//Dispatch methods for EstablishReplicaConsensus
// SendReadRequestToReplicasUtil sends a read confirmation messages from the master to the replicas.
// The replicas send back a check sum
// Otherwise, send back an error.
func (n *Node) SendReadRequestToReplicasUtil(readCheckMsg *pc.ServerMessage, peerRecord *pc.PeerRecord, replyChan chan bool, expectedChecksum []byte) {

	fmt.Printf("Master %d checking read request with replica %d\n", n.myPRecord.Id, peerRecord.Id)

	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	replicaMsg, err := cli.EstablishReplicaConsensus(context.Background(), readCheckMsg)

	if err != nil {
		fmt.Println("SendReadRequestToReplicasUtil: ERROR", err)
		replyChan <- false
		return
	}

	if replicaMsg.Type != pc.ServerMessage_Ack {
		fmt.Println("SendReadRequestToReplicasUtil: NOT OK", replicaMsg.Type)
		replyChan <- false
		return
	}

	fmt.Println(replicaMsg.FileBody.FileContent, expectedChecksum)
	// Check checksum
	if replicaMsg.Type == pc.ServerMessage_Ack && reflect.DeepEqual(replicaMsg.FileBody.FileContent, expectedChecksum) {
		replyChan <- true
	} else {
		replyChan <- false
	}

	fmt.Println("Master read check to replica reply from replica:", replicaMsg.Type)

}

// SendSubRequestToReplicasUtil forwards the subscription request from the master to the replicas.
// The replicas send back an OK.
func (n *Node) SendSubRequestToReplicasUtil(subMsg *pc.ServerMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {

	fmt.Printf("Master %d send sub request %s to replica %d\n", n.myPRecord.Id, subMsg.Type, peerRecord.Id)

	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	replicaMsg, err := cli.EstablishReplicaConsensus(context.Background(), subMsg)

	if err != nil {
		fmt.Println("SendSubRequestToReplicasUtil: ERROR", err)
		replyChan <- false
		return
	}

	replyChan <- replicaMsg.Type == pc.ServerMessage_Ack

	fmt.Println("SendSubRequestToReplicasUtil reply from replica:", replicaMsg.Type)

}

// Forwards locks to Replicas
// Replica adds lock to individual lock files
// Node Lock path
func (n *Node) SendReplicaLocksUtil(lock *pc.ServerMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {
	fmt.Printf("Master %d sends locks to replica %d\n", n.myPRecord.Id, peerRecord.Id)
	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)
	replicaMsg, err := cli.EstablishReplicaConsensus(context.Background(), lock)

	if err != nil {
		fmt.Println("SendReplicaLocksUtil: ERROR", err)
		replyChan <- false
		return
	}

	replyChan <- replicaMsg.Type == pc.ServerMessage_Ack
	fmt.Println("Reply from Replicas regarding locks:", replicaMsg.Type)
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
