package nodecomm

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pc "assignment1/main/protocchubby"
)

const (
	READ_MAX_BYTE_SIZE = 1024
)

func (n *Node) validateReadRequest(CliMsg *pc.ClientMessage) bool {
	return n.validateReadLock() && n.validateFileExists(CliMsg)
}

// TODO: Locking when available
func (n *Node) validateReadLock() bool {
	return true
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

// Returns the checksum of a given file
func (n *Node) getFileChecksum(fullFilePath string) []byte {
	fileToCheck, err := os.Open(fullFilePath)
	if err != nil {
		fmt.Println("getFileChecksum ERROR:", err)
		return []byte{}
	}
	defer fileToCheck.Close()

	checksum := sha256.New()
	if _, err := io.Copy(checksum, fileToCheck); err != nil {
		fmt.Println("getFileChecksum ERROR:", err)
		return []byte{}
	}
	return checksum.Sum(nil)
}

// -----------------------------------------
// |       GRPC READ REPLICA UTILS          |
// ------------------------------------------

// handleReadRequestFromMaster sends back the checksum of the requested file to the master.
func (n *Node) handleReadRequestFromMaster(serverMsg *pc.ServerMessage) *pc.ServerMessage {
	fmt.Printf("> Master requesting checksum for read %s\n", serverMsg.StringMessages)

	// Get Checksum
	localFilePath := filepath.Join(n.nodeDataPath, serverMsg.StringMessages)
	checksum := n.getFileChecksum(localFilePath)

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
