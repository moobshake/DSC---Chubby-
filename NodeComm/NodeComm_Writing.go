package nodecomm

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	pc "assignment1/main/protocchubby"
)

const (
	// eg. temp_file1.txt
	// serves as a temporary place to store writes until replicas
	// acknowledge the writes
	TEMP_PREFIX = "temp_"
)

// TODO: Locking when available
func (n *Node) validateWriteLock() bool {
	return true
}

// This function appends the file information to the end of the file.
// If truncateFile is true, the file is truncated to 0 first before appending the information.
func (n *Node) writeToLocalFile(CliMsg *pc.ClientMessage, truncateFile bool, temp bool) {
	fmt.Println("> server writing to file:", CliMsg.StringMessages, "for client", CliMsg.ClientID, " Temp:", temp)

	var filePath string

	if temp {
		filePath = filepath.Join(n.nodeDataPath, TEMP_PREFIX+CliMsg.StringMessages)
	} else {
		filePath = filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
	}

	// Open file
	file, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("writeToLocalFile ERROR:", err)
	}
	defer file.Close()

	// truncate if needed
	if truncateFile {
		err := os.Truncate(filePath, 0)
		if err != nil {
			fmt.Println("writeToLocalFile TRUNCATE ERROR:", err)
		}
	}

	// write content
	if _, err := file.Write(CliMsg.FileBody.FileContent); err != nil {
		fmt.Println(err)
	}
}

func (n *Node) writeFromTempToLocal(oriFileName string) bool {
	tempFileContent, err := ioutil.ReadFile(filepath.Join(n.nodeDataPath, TEMP_PREFIX+oriFileName))
	if err != nil {
		fmt.Println("writeFromTempToLocal ERROR:", err)
		return false
	}

	err = ioutil.WriteFile(filepath.Join(n.nodeDataPath, oriFileName), tempFileContent, 0644)
	if err != nil {
		fmt.Println("writeFromTempToLocal ERROR:", err)
		return false
	}

	// Delete Temp File
	return n.deleteTempFile(oriFileName)
}

func (n *Node) deleteTempFile(oriFileName string) bool {
	tempFilePath := filepath.Join(n.nodeDataPath, TEMP_PREFIX+oriFileName)
	err := os.Remove(tempFilePath)
	if err != nil {
		fmt.Println("deleteTempFile ERROR:", err)
		return false
	}
	return true
}

// --------------------------------------------
// |       GRPC WRITE STREAMING UTILS          |
// ---------------------------------------------

func (n *Node) handleClientWriteRequest(stream pc.NodeCommListeningService_SendWriteRequestServer, firstMessage *pc.ClientMessage) error {

	if !n.IsMaster() {
		return stream.SendAndClose(n.getRedirectionCliMsg(-1))
	}

	// Store the write requests from the clients so that
	// the master can easily send them to the replicas
	writeRequestBuffers := make([]*pc.ClientMessage, 0)

	// This is the first message from the client that should
	// contain a valid write lock.
	writeRequestMessage := firstMessage
	writeRequestBuffers = append(writeRequestBuffers, writeRequestMessage)

	// Validate write lock
	// TODO(Hannah): change to appropriate function
	if n.validateWriteLock() {
		n.writeToLocalFile(writeRequestMessage, true, true)

		// Keep listening for more messages from the client in case
		// the file is very big.
		for {
			writeRequestMessage, err := stream.Recv()
			if err == io.EOF {
				// Make sure that the majority of replicas give their OK to writing
				if n.SendWriteRequestToReplicas(writeRequestBuffers) {
					// Majority of replicas gave their ok, write from temp to local file
					if n.writeFromTempToLocal(writeRequestBuffers[0].StringMessages) {
						// publish file modification event
						go n.PublishFileContentModification(writeRequestBuffers[0].StringMessages, writeRequestBuffers[0].ClientAddress)
						return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_FileWrite})
					} else {
						n.deleteTempFile(writeRequestBuffers[0].StringMessages)
						return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_Error})
					}
				} else {
					// No majoirty --> ERROR
					return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_Error})
				}
			}
			if err != nil {
				return err
			}

			n.writeToLocalFile(writeRequestMessage, false, true)
			writeRequestBuffers = append(writeRequestBuffers, writeRequestMessage)
		}
	} else {
		return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_InvalidLock})
	}
}

func (n *Node) handleMasterToReplicatWriteRequest(stream pc.NodeCommListeningService_SendWriteRequestServer, firstMessage *pc.ClientMessage) error {
	// This is the first message from the client that should
	// contain a valid write lock.
	writeRequestMessage := firstMessage

	n.writeToLocalFile(writeRequestMessage, true, false)

	// Keep listening for more messages from the master in case
	// the file is very big.
	for {
		writeRequestMessage, err := stream.Recv()
		if err == io.EOF {
			// Make sure that the majority of replicas give their OK to writing
			return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_Ack})
		}
		if err != nil {
			return err
		}
		n.writeToLocalFile(writeRequestMessage, false, false)
	}

}

// Write requests should be acknowledged by the majority of replicas
// This function shall wait for the majoirty of replicas to return OK
// After a timeout, if majority do not agree, timeout and fail
func (n *Node) SendWriteRequestToReplicas(writeMsgBuffer []*pc.ClientMessage) bool {
	// The go routines for the writing functions will return if the replica writes were successful
	replicaReplyChan := make(chan bool, 1)
	// Count the number of successful replica writes
	countAck := 0

	// Timeout setup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, peerRecord := range n.peerRecords {
		go n.SendWriteRequestToReplicasUtil(writeMsgBuffer, peerRecord, replicaReplyChan)
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

// sendClientWriteRequest sends a stream of messages from the master to the replicas.
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
