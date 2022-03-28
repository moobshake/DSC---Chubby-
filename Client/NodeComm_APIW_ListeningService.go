package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

//DispatchClientMessage sends a client message
func (c *Client) DispatchClientMessage(destPRec *pc.PeerRecord, CliMsg *pc.ClientMessage) *pc.ClientMessage {
	if destPRec == nil {
		fmt.Println("Error dispatching client message because no destination Peer Record was provided.")
		return nil
	}
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	cConn := pc.NewNodeCommListeningServiceClient(conn)
	response, err := cConn.SendClientMessage(context.Background(), CliMsg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}
	return response
}

//Msg from Jia Wei: Does the DispatchReadRequest need to return something for methods that call it to use it?
//Or does the way the locks work mean that just a simple call with no return is necessary?

// DispatchReadRequest sends a read request to the server.
// The server streams back the file content if the client has a valid lock.
// Returns if the read request was successful
func (c *Client) DispatchReadRequest(readFileName string) {
	fmt.Printf("Client %d creating Read Request\n", c.ClientID)

	readLock := c.getValidLocalReadLock(readFileName)

	cliMsg := pc.ClientMessage{
		ClientID: int32(c.ClientID),
		Type:     pc.ClientMessage_FileRead,
		// The name of the file to read
		StringMessages: readFileName,
		Lock:           readLock,
	}

	conn, err := connectTo(c.MasterAdd.Address, c.MasterAdd.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommListeningServiceClient(conn)

	stream, err := cli.SendReadRequest(context.Background(), &cliMsg)

	if err != nil {
		fmt.Println("DispatchReadRequest: ERROR", err)
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
			fmt.Println("DispatchReadRequest: ERROR", err)
		}

		if fileContent.Type == pc.FileBodyMessage_Error {
			fmt.Println("Server returned an error for file reading:", cliMsg.StringMessages)
			break
		}

		if fileContent.Type == pc.FileBodyMessage_InvalidLock {
			fmt.Println("Server sent back:", fileContent.Type)
			break
		}

		c.writeToCache(fileContent, truncateFile)
		// Append the rest of the content to the file
		truncateFile = false
		fmt.Println("Client received a successful read block")
	}
}

// sendClientWriteRequest sends a stream of messages to the server.
// The client sends the whole file that it wishes to modify to the server.
// The client should send a valid write lock.
// If the lock is invalid the client stops trying to send the file.
// For demonstration purposes, this function might modify the file first before
// sending it to the server.
func (c *Client) sendClientWriteRequest(writeFileName string, shouldModifyFile bool) {
	writeLock := c.getValidLocalWriteLock(writeFileName)

	// create stream for message sending
	conn, err := connectTo(c.MasterAdd.Address, c.MasterAdd.Port)
	if err != nil {
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommListeningServiceClient(conn)

	stream, err := cli.SendWriteRequest(context.Background())
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
		fileContent := pc.FileBodyMessage{
			Type:        pc.FileBodyMessage_WriteMode,
			FileName:    writeFileName,
			FileContent: buffer[:numBytes],
		}

		cliMsg := pc.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     pc.ClientMessage_FileWrite,
			// The name of the file to write
			StringMessages: writeFileName,
			FileBody:       &fileContent,
			Lock:           writeLock,
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
