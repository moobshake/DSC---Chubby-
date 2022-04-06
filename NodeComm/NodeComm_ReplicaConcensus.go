package nodecomm

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"time"

	pc "assignment1/main/protocchubby"
)

// This function shall wait for the majoirty of replicas to return OK
// After a timeout, if majority do not agree, timeout and fail
// Note that write requests are handled in SendWriteRequestToReplicas
// Note: This function assumes that there are at least 2 Servers - Master + 1 Replica Alive
// Otherwise, it will always return false
func (n *Node) SendRequestToReplicas(serverMessage *pc.ServerMessage) bool {
	// The go routines for the util functions will return true if the replica replications were successful
	replicaReplyChan := make(chan bool, 1)
	// Count the number of successful replica writes
	countAck := 0

	// Timeout setup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	switch serverMessage.Type {
	case pc.ServerMessage_ReplicaReadCheck:
		for _, peerRecord := range n.peerRecords {
			checksum := n.getFileChecksum(filepath.Join(n.nodeDataPath, serverMessage.StringMessages))
			go n.SendReadRequestToReplicasUtil(serverMessage, peerRecord, replicaReplyChan, checksum)
		}
	case pc.ServerMessage_SubscribeFileModification, pc.ServerMessage_SubscribeLockConflict,
		pc.ServerMessage_SubscribeMasterFailover, pc.ServerMessage_SubscribeLockAquisition:
		for _, peerRecord := range n.peerRecords {
			go n.SendSubRequestToReplicasUtil(serverMessage, peerRecord, replicaReplyChan)
		}
	// Client requested for lock
	// TODO: YH add lock cases
	case pc.ServerMessage_ReqLock:
		for _, peerRecord := range n.peerRecords {
			go n.SendReplicaLocksUtil(serverMessage, peerRecord, replicaReplyChan)
		}

	}

	// Wait for all go-routines or timeout
	for i := 0; i < len(n.peerRecords); i++ {
		select {
		case <-ctx.Done():
			// this is usually becase of timeout
			fmt.Println("SendRequestToReplicas Context Error:", ctx.Err())
		case reply := <-replicaReplyChan:
			fmt.Println("Replica Write reply:", reply)
			if reply {
				countAck++
			}
		}
	}

	num_majority := len(n.peerRecords)/2 + 1
	fmt.Println("Num replicas:", len(n.peerRecords), "majority num:", num_majority, "num acks:", countAck)
	return countAck >= num_majority
}

// This function shall wait for the majoirty of replicas to return OK
// After a timeout, if majority do not agree, timeout and fail
func (n *Node) SendWriteRequestToReplicas(CliMsgBuffer []*pc.ClientMessage) bool {
	// The go routines for the writing functions will return if the replica writes were successful
	replicaReplyChan := make(chan bool, 1)
	// Count the number of successful replica writes
	countAck := 0

	// Timeout setup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, peerRecord := range n.peerRecords {
		go n.SendWriteRequestToReplicasUtil(CliMsgBuffer, peerRecord, replicaReplyChan)
	}

	// Wait for all go-routines or timeout
	for i := 0; i < len(n.peerRecords); i++ {
		select {
		case <-ctx.Done():
			fmt.Println("SendWriteRequestToReplicas Context Error:", ctx.Err())
		case reply := <-replicaReplyChan:
			fmt.Println("Replica Write reply:", reply)
			if reply {
				countAck++
			}
		}
	}

	num_majority := len(n.peerRecords)/2 + 1
	fmt.Println("Num replicas:", len(n.peerRecords), "majority num:", num_majority, "num acks:", countAck)
	return countAck >= num_majority
}

// Write requests should be acknowledged by the majority of replicas
// SendWriteRequestToReplicasUtil sends a stream of messages from the master to the replicas.
// The replicas should save the file to their local copies and return an OK if they are done
// Otherwise, send back an error.
func (n *Node) SendWriteRequestToReplicasUtil(writeMsgBuffer []*pc.ClientMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {
	// create stream for message sending
	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommListeningServiceClient(conn)

	stream, err := cli.SendWriteRequest(context.Background())
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
