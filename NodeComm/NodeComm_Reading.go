package nodecomm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pc "assignment1/main/protocchubby"
)

const (
	READ_MAX_BYTE_SIZE = 1024
)

func (n *Node) validateReadRequest(CliMsg *pc.ClientMessage) bool {
	return n.validateReadLock(int(CliMsg.ClientID), CliMsg.Lock.Sequencer) && n.validateFileExists(CliMsg)
}

func (n *Node) validateReadLock(id int, sequencer string) bool {
	filename := strings.Split(sequencer, ",")[0]

	// get lock file
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
	}
	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}
	// read lock empty, means it expired
	if len(l.Read) == 0 {
		// check for write lock. if write lock is valid, then it is okay
		if len(l.Write) != 0 {
			for i := range l.Write {
				if i == id && l.Write[i].Sequence == sequencer {
					fmt.Println("Valid write lock, allow client to read too.")
					return true
				}
			}
		}
	} else {
		for i := range l.Read {
			// if id is found in l.read and sequencer is the same
			if i == id && l.Read[i].Sequence == sequencer {
				fmt.Println("Valid read lock")
				return true
			}
		}
	}

	return false
}

// If the file exists, store it in local memory
func (n *Node) validateFileExists(CliMsg *pc.ClientMessage) bool {
	//local file storage path
	localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)

	// Check if file exists in local dir
	if _, err := os.Stat(localFilePath); err == nil {
		// exists
		return true
	} else {
		return false
	}
}

// -----------------------------------------
// |       GRPC READ REPLICA UTILS          |
// ------------------------------------------

// handleReadRequestFromMaster sends back the checksum of the requested file to the master.
func (n *Node) handleReadRequestFromMaster(serverMsg *pc.ServerMessage) *pc.ServerMessage {
	fmt.Printf("> Master requesting checksum for read %s\n", serverMsg.StringMessages)

	// Get Checksum
	localFilePath := filepath.Join(n.nodeDataPath, serverMsg.StringMessages)
	checksum := getFileChecksum(localFilePath)

	// Checksum in a byte array, send it as a file content
	fileContent := pc.FileBodyMessage{
		Type:        pc.FileBodyMessage_ReadMode,
		FileName:    serverMsg.StringMessages,
		FileContent: checksum,
	}

	serverMessage := pc.ServerMessage{
		Type:     pc.ServerMessage_Ack,
		FileBody: &fileContent,
	}

	return &serverMessage
}
