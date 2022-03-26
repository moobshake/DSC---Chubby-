package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	NC "assignment1/main/NodeComm"
)

const (
	APPENDING_MODIFICATION = "\nThis is a new line"
	READ_MAX_BYTE_SIZE     = 1024
)

// Sends a stream of messages to the server.
// The client sends the whole file that it wishes to modify to the server.
// The client should send a valid write lock.
// If the lock is invalid the client stops trying to send the file.
// For demonstration purposes, this function might modify the file first before
// sending it to the server.
// TODO(Hannah): Send valid write lock
func (c *Client) sendClientWriteRequest(writeFileName string, shouldModifyFile bool) {
	// create stream for messaging sending
	conn, err := connectTo(c.MasterAdd.Address, c.MasterAdd.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := NC.NewNodeCommServiceClient(conn)

	stream, err := cli.RequestWriteFile(context.Background())
	if err != nil {
		fmt.Println("CLIENT FILE WRITE REQUEST ERROR:", err)
	}

	// If we need to modify or create the file first, do this
	if shouldModifyFile {
		c.modifyFileForDemo(writeFileName)
	}

	// Retrive the file to send
	cacheFilePath := filepath.Join(c.ClientCacheFilePath, writeFileName)
	file, err := os.Open(cacheFilePath)
	if err != nil {
		fmt.Println("CLIENT FILE WRITE REQUEST ERROR:", err)
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

		// The file content to embed in the request
		fileContent := NC.FileBodyMessage{
			Type:        NC.FileBodyMessage_WriteMode,
			FileName:    writeFileName,
			FileContent: buffer[:numBytes],
		}

		cliMsg := NC.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     NC.ClientMessage_FileWrite,
			// The name of the file to write
			StringMessages: writeFileName,
			FileBody:       &fileContent,
			Lock:           &NC.LockMessage{},
		}

		if err := stream.Send(&cliMsg); err != nil {
			fmt.Println("CLIENT FILE WRITE REQUEST ERROR:", err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("CLIENT FILE WRITE REQUEST ERROR:", err)
	}
	fmt.Println("Client write request reply from server:", reply.Type)
}

// Append a single line to the file that should be sent to the server.
// This way we can see that the file has been changed at the server.
// If the file does not exist, create it.
func (c *Client) modifyFileForDemo(writeFileName string) {
	// Create cache directory if not exists
	err := os.MkdirAll(c.ClientCacheFilePath, os.ModePerm)
	if err != nil {
		fmt.Println("modifyFileForDemo ERROR:", err)
	}

	// actual full file path
	cacheFilePath := filepath.Join(c.ClientCacheFilePath, writeFileName)

	// Open file
	cache_file, err := os.OpenFile(cacheFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("modifyFileForDemo ERROR:", err)
	}
	defer cache_file.Close()

	// write content
	if _, err := cache_file.WriteString(APPENDING_MODIFICATION); err != nil {
		fmt.Println(err)
	}
}
