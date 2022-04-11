package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	lockStatus := c.isLockExpire(writeFileName)
	if lockStatus == 2 {
		lock := c.Locks[writeFileName]
		if lock.lockType == WRITE_CLI {
			fmt.Println("Retrived write lock:", lock.sequencer)
			// convert valid lock to lock message
			return &pc.LockMessage{Sequencer: lock.sequencer}
		} else if lock.lockType == READ_CLI {
			fmt.Println("Currently holding read lock")
			return &pc.LockMessage{}
		}
	} else if lockStatus == 0 {
		c.ClientRequest(REQ_LOCK, WRITE_CLI, writeFileName)
		time.Sleep(time.Second * 1)
		if c.isLockExpire(writeFileName) == 2 {
			lock := c.Locks[writeFileName]
			if lock.lockType == WRITE_CLI {
				fmt.Println("Retrived write lock:", lock.sequencer)
				// convert valid lock to lock message
				return &pc.LockMessage{Sequencer: lock.sequencer}
			}
		}
	}

	// no valid lock to return
	fmt.Println("Unable to retrieve a valid write lock")
	return &pc.LockMessage{}
}
