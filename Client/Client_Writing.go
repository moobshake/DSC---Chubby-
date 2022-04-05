package client

import (
	"fmt"
	"os"
	"path/filepath"

	pc "assignment1/main/protocchubby"
)

const (
	APPENDING_MODIFICATION = "\nThis is a new line"
	READ_MAX_BYTE_SIZE     = 1024
)

// modifyFileForDemo appends a single line to the file that should be sent to the server.
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

// getValidLocalWriteLock checks if the client has the appropraite valid write lock.
// If the client does not have the lock, request for it first.
// Returns a valid lock, otherwise, return an empty lock.
// TODO(Hannah): Get valid write lock
func (c *Client) getValidLocalWriteLock(writeFileName string) *pc.LockMessage {
	// check if the client already has any lock for this file
	if _, ok := c.Locks[writeFileName]; !ok {
		// Client does not currently have the lock, request for it.
		c.ClientRequest(REQ_LOCK, WRITE_CLI, writeFileName)
	}

	// a lock was sucessfully retrived, check type
	if writeLock, ok := c.Locks[writeFileName]; ok {
		if writeLock.lockType == WRITE_CLI {
			// check if lock has expired ________________________________ Check with hannah
			if c.LockCheckExpire(writeFileName) {
				fmt.Println("Retrived write lock:", writeLock.sequencer)
				// convert valid lock to lock message
				return &pc.LockMessage{Sequencer: writeLock.sequencer}
			}
		}
	}

	// no valid lock to return
	fmt.Println("Unable to retrieve a valid write lock")
	return &pc.LockMessage{}
}
