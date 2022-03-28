package nodecomm

import (
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
