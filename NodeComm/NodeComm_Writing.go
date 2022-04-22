package nodecomm

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pc "assignment1/main/protocchubby"
)

const (
	// eg. temp_file1.txt
	// serves as a temporary place to store writes until replicas
	// acknowledge the writes
	TEMP_PREFIX = "temp_"
)

func (n *Node) validateWriteLock(id int, sequencer string) bool {
	filename := strings.Split(sequencer, ",")[0]

	// get lock file
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
		return false
	}
	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}

	// if not empty, else return false
	if len(l.Write) != 0 {
		for i := range l.Write {
			if i == id && l.Write[i].Sequence == sequencer {
				fmt.Println("Valid write lock")
				return true
			}
		}
	}

	return false
}

// This function appends the file information to the end of the file.
// If truncateFile is true, the file is truncated to 0 first before appending the information.
func (n *Node) writeToLocalFile(fileBody *pc.FileBodyMessage, truncateFile bool) {
	fmt.Println("> server writing to file:")

	// Get the correct file path for this machine os
	getCorrectFilePath(&fileBody.FileName)

	fullFilePath := filepath.Join(n.nodeRootPath, fileBody.FileName)
	fmt.Println("writeToLocalFile:", fileBody.FileName, fullFilePath, "root", n.nodeRootPath)

	// Open file
	file, err := os.OpenFile(fullFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("writeToLocalFile ERROR:", err)
	}
	defer file.Close()

	// truncate if needed
	if truncateFile {
		err := os.Truncate(fullFilePath, 0)
		if err != nil {
			fmt.Println("writeToLocalFile TRUNCATE ERROR:", err)
		}
	}

	// write content
	if _, err := file.Write(fileBody.FileContent); err != nil {
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
	if n.validateWriteLock(int(firstMessage.ClientID), firstMessage.Lock.Sequencer) {
		// publish file modification event to invalidate other client caches
		go n.PublishFileContentModification(writeRequestBuffers[0].StringMessages, writeRequestBuffers[0].ClientAddress)

		writeRequestMessage.FileBody.FileName = filepath.Join(LOCAL_DATA_DIR_PREFIX, TEMP_PREFIX+writeRequestMessage.StringMessages)
		n.writeToLocalFile(writeRequestMessage.FileBody, true)

		// Keep listening for more messages from the client in case
		// the file is very big.
		for {
			writeRequestMessage, err := stream.Recv()
			if err == io.EOF {
				// Make sure that the majority of replicas give their OK to writing
				if n.SendWriteRequestToReplicas(writeRequestBuffers) {
					// Majority of replicas gave their ok, write from temp to local file
					if n.writeFromTempToLocal(writeRequestBuffers[0].StringMessages) {
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

			writeRequestMessage.FileBody.FileName = filepath.Join(LOCAL_DATA_DIR_PREFIX, TEMP_PREFIX+writeRequestMessage.StringMessages)
			n.writeToLocalFile(writeRequestMessage.FileBody, false)
			writeRequestBuffers = append(writeRequestBuffers, writeRequestMessage)
		}
	} else {
		return stream.SendAndClose(&pc.ClientMessage{Type: pc.ClientMessage_InvalidLock})
	}
}

func (n *Node) handleMasterToReplicatWriteRequest(stream pc.NodeCommPeerService_SendWriteForwardServer, firstMessage *pc.ServerMessage) error {
	writeRequestMessage := firstMessage

	// fullFilePath := ""
	// if writeRequestMessage.Type == pc.ServerMessage_ReplicaWriteData {
	// 	fullFilePath = filepath.Join(n.nodeDataPath, writeRequestMessage.FileBody.FileName)
	// } else {
	// 	fullFilePath = filepath.Join(n.nodeLockPath, writeRequestMessage.FileBody.FileName)
	// }

	n.writeToLocalFile(writeRequestMessage.FileBody, true)

	// Keep listening for more messages from the master in case
	// the file is very big.
	for {
		writeRequestMessage, err := stream.Recv()
		if err == io.EOF {
			delete(n.outstandingFiles, firstMessage.FileBody.FileName)
			// Make sure that the majority of replicas give their OK to writing
			return stream.SendAndClose(&pc.ServerMessage{Type: pc.ServerMessage_Ack})
		}
		if err != nil {
			return err
		}
		n.writeToLocalFile(writeRequestMessage.FileBody, false)
	}
	// Update list of oustanding files
}
