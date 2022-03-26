package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	NC "assignment1/main/NodeComm"
)

// Sends a read request to the server.
// The server streams back the file content if the client has a valid lock.
// Returns if the read request was successful
func (c *Client) sendClientReadRequest(readFileName string) {
	fmt.Printf("Client %d creating Read Request\n", c.ClientID)

	read_lock := c.getValidLocalReadLock(readFileName)

	cliMsg := NC.ClientMessage{
		ClientID: int32(c.ClientID),
		Type:     NC.ClientMessage_FileRead,
		// The name of the file to read
		StringMessages: readFileName,
		Lock:           read_lock,
	}

	conn, err := connectTo(c.MasterAdd.Address, c.MasterAdd.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := NC.NewNodeCommServiceClient(conn)

	stream, err := cli.RequestReadFile(context.Background(), &cliMsg)

	if err != nil {
		fmt.Println("sendClientReadRequest: ERROR", err)
	}

	// Always truncate the cache file first
	truncateFile := true

	for {
		fileContent, err := stream.Recv()
		if err == io.EOF {
			// Stream has ended
			break
		}
		if err != nil {
			fmt.Println("sendClientReadRequest: ERROR", err)
		}

		if fileContent.Type == NC.FileBodyMessage_Error {
			fmt.Println("Server returned an error for file reading:", cliMsg.StringMessages)
			break
		}

		if fileContent.Type == NC.FileBodyMessage_InvalidLock {
			fmt.Println("Server sent back:", fileContent.Type)
			break
		}

		c.writeToCache(fileContent, truncateFile)
		// Append the rest of the content to the file
		truncateFile = false
		fmt.Println("Client received a successful read block")
	}

}

// This function appends the file information to the end of the file.
// If truncateFile is true, the file is truncated to 0 first before appending the information.
func (c *Client) writeToCache(fileContentMessage *NC.FileBodyMessage, truncateFile bool) {
	// Create cache directory if not exists
	err := os.MkdirAll(c.ClientCacheFilePath, os.ModePerm)
	if err != nil {
		fmt.Println("writeToCache ERROR:", err)
	}

	// actual full file path
	cacheFilePath := filepath.Join(c.ClientCacheFilePath, fileContentMessage.FileName)

	// Open file
	cache_file, err := os.OpenFile(cacheFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("writeToCache ERROR:", err)
	}
	defer cache_file.Close()

	// truncate if needed
	if truncateFile {
		err := os.Truncate(cacheFilePath, 0)
		if err != nil {
			fmt.Println("writeToCache TRUNCATE ERROR:", err)
		}
	}

	// write content
	if _, err := cache_file.Write(fileContentMessage.FileContent); err != nil {
		fmt.Println(err)
	}
}

// Check if the client has the appropraite valid read lock.
// If the client does not have the lock, request for it first.
// Returns a valid lock, otherwise, return an empty lock.
// TODO(Hannah): Get valid read lock
func (c *Client) getValidLocalReadLock(readFileName string) *NC.LockMessage {
	// check if the client already has any lock for this file
	if _, ok := c.Locks[readFileName]; !ok {
		// Client does not currently have the lock, request for it.
		c.ClientRequest(REQ_LOCK, READ_CLI, readFileName)
	}

	// a lock was sucessfully retrived, check type
	if read_lock, ok := c.Locks[readFileName]; ok {
		if read_lock.l_type == READ_CLI {
			fmt.Println("Retrived read lock:", read_lock.sequencer)
			// convert valid lock to lock message
			return &NC.LockMessage{Sequencer: read_lock.sequencer}
		}
	}

	// no valid lock to return
	fmt.Println("Unable to retrieve a valid read lock")
	return &NC.LockMessage{}
}
