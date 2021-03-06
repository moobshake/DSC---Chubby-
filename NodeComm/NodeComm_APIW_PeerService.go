package nodecomm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
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

// DispatchFileToReplica sends a stream of messages to a replica.
func (n *Node) DispatchFileToReplica(dPRec *pc.PeerRecord, filePath string) {

	// Get the correct file path for this machine os
	getCorrectFilePath(&filePath)

	fullFilePath := filepath.Join(n.nodeRootPath, filePath)

	// create stream for message sending
	conn, err := connectTo(dPRec.Address, dPRec.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	stream, err := cli.SendWriteForward(context.Background())

	if err != nil {
		fmt.Println("Error sending file to replica.", err)
		return
	}

	// Retrive the file to send
	file, err := os.Open(fullFilePath)
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

		// tokenisedFilePath := strings.Split(filePath, "/")
		// fileName := tokenisedFilePath[len(tokenisedFilePath)-1]

		// The file content to embed in the request
		fileContent := pc.FileBodyMessage{
			Type:        pc.FileBodyMessage_WriteMode,
			FileName:    filePath,
			FileContent: buffer[:numBytes],
		}

		serverMsg := pc.ServerMessage{
			// ClientID: int32(n.myPRecord.Id),
			Type: pc.ServerMessage_ReplicaWriteData,
			// The name of the file to write
			StringMessages: filePath,
			FileBody:       &fileContent,
			PeerRecord:     n.myPRecord,
		}

		if err := stream.Send(&serverMsg); err != nil {
			fmt.Println("Error when Master tried to send file to replica:", err)
		}
	}
	stream.CloseAndRecv()
}

// Write requests should be acknowledged by the majority of replicas
// SendWriteRequestToReplicasUtil sends a stream of messages from the master to the replicas.
// The replicas should save the file to their local copies and return an OK if they are done
// Otherwise, send back an error.
// This function cannot be broken up further without overcomplicating things.
func (n *Node) DispatchWriteRequestToReplicasUtil(writeMsgBuffer []*pc.ClientMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {
	// create stream for message sending
	// create stream for message sending
	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommPeerServiceClient(conn)

	stream, err := cli.SendWriteForward(context.Background())

	if err != nil {
		replyChan <- false
		return
	}

	for _, cliWriteMsg := range writeMsgBuffer {
		// The file content to embed in the request
		serverMsg := pc.ServerMessage{
			Type:     pc.ServerMessage_ReplicaWriteData,
			FileBody: cliWriteMsg.FileBody,
		}
		serverMsg.FileBody.FileName = filepath.Join(LOCAL_DATA_DIR_PREFIX, cliWriteMsg.StringMessages)

		if err := stream.Send(&serverMsg); err != nil {
			fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		replyChan <- false
		return
	}

	replyChan <- reply.Type == pc.ServerMessage_Ack

	fmt.Println("Master write to replica reply from replica:", reply.Type)
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
