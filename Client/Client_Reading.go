package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	pc "assignment1/main/protocchubby"
)

// This function appends the file information to the end of the file.
// If truncateFile is true, the file is truncated to 0 first before appending the information.
func (c *Client) writeToCache(fileContentMessage *pc.FileBodyMessage, truncateFile bool) {
	// Create cache directory if not exists
	err := os.MkdirAll(c.ClientCacheFilePath, os.ModePerm)
	if err != nil {
		fmt.Println("writeToCache ERROR:", err)
	}

	// actual full file path
	cacheFilePath := filepath.Join(c.ClientCacheFilePath, fileContentMessage.FileName)

	// Open file
	cacheFile, err := os.OpenFile(cacheFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("writeToCache ERROR:", err)
	}
	defer cacheFile.Close()

	// truncate if needed
	if truncateFile {
		err := os.Truncate(cacheFilePath, 0)
		if err != nil {
			fmt.Println("writeToCache TRUNCATE ERROR:", err)
		}
	}

	// write content
	if _, err := cacheFile.Write(fileContentMessage.FileContent); err != nil {
		fmt.Println(err)
	}
}

// Check if the client has the appropraite valid read lock.
// If the client does not have the lock, request for it first.
// Returns a valid lock, otherwise, return an empty lock.
// TODO(Hannah): Get valid read lock
func (c *Client) getValidLocalReadLock(readFileName string) *pc.LockMessage {
	lockStatus := c.isLockExpire(readFileName)
	if lockStatus == 2 {
		lock := c.Locks[readFileName]
		if lock.lockType == READ_CLI {
			fmt.Println("Retrived read lock:", lock.sequencer)
			// convert valid lock to lock message
			return &pc.LockMessage{Sequencer: lock.sequencer, Type: pc.LockMessage_ReadLock, LockDelay: int32(lock.lockDelay), TimeStamp: lock.timestamp.String()}
		} else if lock.lockType == WRITE_CLI {
			fmt.Println("Retrived write lock, can be used as read lock:", lock.sequencer)
			// convert valid lock to lock message
			return &pc.LockMessage{Sequencer: lock.sequencer, Type: pc.LockMessage_WriteLock, LockDelay: int32(lock.lockDelay), TimeStamp: lock.timestamp.String()}
		}
	} else if lockStatus == 0 {
		c.ClientRequest(REQ_LOCK, READ_CLI, readFileName)
		time.Sleep(time.Second * 1)
		if c.isLockExpire(readFileName) == 2 {
			lock := c.Locks[readFileName]
			if lock.lockType == READ_CLI {
				fmt.Println("Retrived read lock:", lock.sequencer)
				// convert valid lock to lock message
				return &pc.LockMessage{Sequencer: lock.sequencer, Type: pc.LockMessage_ReadLock, LockDelay: int32(lock.lockDelay), TimeStamp: lock.timestamp.String()}
			} else if lock.lockType == WRITE_CLI {
				fmt.Println("Retrived write lock, can be used as read lock:", lock.sequencer)
				// convert valid lock to lock message
				return &pc.LockMessage{Sequencer: lock.sequencer, Type: pc.LockMessage_WriteLock, LockDelay: int32(lock.lockDelay), TimeStamp: lock.timestamp.String()}
			}
		}
	}

	// no valid lock to return
	fmt.Println("Unable to retrieve a valid read lock")
	return &pc.LockMessage{}
}
