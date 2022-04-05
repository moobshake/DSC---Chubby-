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

// -------------------------------------------
// |       GRPC READ STREAMING UTILS          |
// --------------------------------------------

// handleReadRequestFromClient handles the read request from the client directly
// It streams the required file from local data to the client in batches
func (n *Node) handleReadRequestFromClient(CliMsg *pc.ClientMessage, stream pc.NodeCommListeningService_SendReadRequestServer) error {
	fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)

	if !n.IsMaster() {
		// Return a master redirection message
		cliMsg := n.getRedirectionCliMsg(CliMsg.ClientID)
		if err := stream.Send(cliMsg); err != nil {
			return err
		}

		return nil
	}

	if n.validateReadRequest(CliMsg) {
		// Check if the file is consistent across the majority of replicas
		if !n.SendRequestToReplicas([]*pc.ClientMessage{CliMsg}, pc.ClientMessage_FileRead) {
			// Return an Error
			cliMsg := pc.ClientMessage{
				Type:           pc.ClientMessage_Error,
				StringMessages: "Majority of replicas do not agree on the read file.",
			}
			if err := stream.Send(&cliMsg); err != nil {
				return err
			}
			return nil
		}

		// Get the file from the local dir in batches
		localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
		file, err := os.Open(localFilePath)
		if err != nil {
			fmt.Println("CLIENT FILE READ REQUEST ERROR:", err)
		}
		defer file.Close()

		buffer := make([]byte, READ_MAX_BYTE_SIZE)

		for {
			numBytes, err := file.Read(buffer)

			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
			fileContent := pc.FileBodyMessage{
				Type:        pc.FileBodyMessage_ReadMode,
				FileName:    CliMsg.StringMessages,
				FileContent: buffer[:numBytes],
			}

			cliMsg := pc.ClientMessage{
				Type:     pc.ClientMessage_FileRead,
				FileBody: &fileContent,
			}
			if err := stream.Send(&cliMsg); err != nil {
				return err
			}
		}

		return nil
	} else {
		// Return an invalid lock error
		cliMsg := pc.ClientMessage{
			Type: pc.ClientMessage_InvalidLock,
		}
		if err := stream.Send(&cliMsg); err != nil {
			return err
		}
		return nil
	}

}

// handleReadRequestFromMaster sends back the checksum of the requested file.
func (n *Node) handleReadRequestFromMaster(CliMsg *pc.ClientMessage, stream pc.NodeCommListeningService_SendReadRequestServer) error {
	fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)

	// Get the file from the local dir in batches
	localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
	checksum := n.getFileChecksum(localFilePath)

	fileContent := pc.FileBodyMessage{
		Type:        pc.FileBodyMessage_ReadMode,
		FileName:    CliMsg.StringMessages,
		FileContent: checksum,
	}

	cliMsg := pc.ClientMessage{
		Type:     pc.ClientMessage_Ack,
		FileBody: &fileContent,
	}

	if err := stream.Send(&cliMsg); err != nil {
		return err
	}

	return nil
}
