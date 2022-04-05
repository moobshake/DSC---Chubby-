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
func (n *Node) SendRequestToReplicas(CliMsgBuffer []*pc.ClientMessage, requestType pc.ClientMessage_MessageType) bool {
	// The go routines for the writing functions will return if the replica writes were successful
	replicaReplyChan := make(chan bool, 1)
	// Count the number of successful replica writes
	countAck := 0

	// Timeout setup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	switch requestType {
	case pc.ClientMessage_FileWrite:
		for _, peerRecord := range n.peerRecords {
			go n.SendWriteRequestToReplicasUtil(CliMsgBuffer, peerRecord, replicaReplyChan)
		}
	case pc.ClientMessage_FileRead:
		for _, peerRecord := range n.peerRecords {
			checksum := n.getFileChecksum(filepath.Join(n.nodeDataPath, CliMsgBuffer[0].StringMessages))
			go n.SendReadRequestToReplicasUtil(CliMsgBuffer[0], peerRecord, replicaReplyChan, checksum)
		}
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

//TODO: Use this function
func (n *Node) SendReadRequestToReplicasUtil(readCheckMsg *pc.ClientMessage, peerRecord *pc.PeerRecord, replyChan chan bool, expectedChecksum []byte) {

	fmt.Printf("Master %d checking read request with replica %d\n", n.myPRecord.Id, peerRecord.Id)

	readCheckMsg.Type = pc.ClientMessage_ReplicaReadCheck

	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommListeningServiceClient(conn)

	stream, err := cli.SendReadRequest(context.Background(), readCheckMsg)

	if err != nil {
		fmt.Println("SendReadRequestToReplicasUtil: ERROR", err)
		replyChan <- false
		return
	}

	replicaMsg, err := stream.Recv()
	if err != nil {
		fmt.Println("SendReadRequestToReplicasUtil: ERROR", err)
		replyChan <- false
		return
	}

	if replicaMsg.Type != pc.ClientMessage_Ack {
		fmt.Println("SendReadRequestToReplicasUtil: NOT OK", replicaMsg.Type)

		replyChan <- false
		return
	}

	fmt.Println(replicaMsg.FileBody.FileContent, expectedChecksum)
	// Check checksum
	if replicaMsg.Type == pc.ClientMessage_Ack && reflect.DeepEqual(replicaMsg.FileBody.FileContent, expectedChecksum) {
		replyChan <- true
	} else {
		replyChan <- false
	}

	fmt.Println("Master read check to replica reply from replica:", replicaMsg.Type)

}
