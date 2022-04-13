package client

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pc "assignment1/main/protocchubby"
)

const (
	READ_ERROR = "READ_ERROR"
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
		if lock.lockType == READ_CLI || lock.lockType == WRITE_CLI {
			fmt.Println("Retrived read lock:", lock.sequencer)
			// convert valid lock to lock message
			return &pc.LockMessage{Sequencer: lock.sequencer}
		}
	} else if lockStatus == 0 {
		c.ClientRequest(REQ_LOCK, READ_CLI, readFileName)
		time.Sleep(time.Second * 1)
		if c.isLockExpire(readFileName) == 2 {
			lock := c.Locks[readFileName]
			if lock.lockType == READ_CLI || lock.lockType == WRITE_CLI {
				fmt.Println("Retrived read lock:", lock.sequencer)
				// convert valid lock to lock message
				return &pc.LockMessage{Sequencer: lock.sequencer}
			}
		}
	}

	// no valid lock to return
	fmt.Println("Unable to retrieve a valid read lock")
	return &pc.LockMessage{}
}

// Returns a string of the first numLines lines of a file.
func (c *Client) readFile(readFileName string, numLines int) string {
	// Check validity of file being read.
	// Either no valid cache, or have valid cache but lock expired
	if !c.ClientCacheValidation[readFileName] ||
		(c.ClientCacheValidation[readFileName] && c.isLockExpire(readFileName) != 2) {
		c.ClientCacheValidation[readFileName] = false
		fmt.Println(readFileName, "has no valid cache. ")
		return READ_ERROR
	}

	output := ""

	cacheFilePath := filepath.Join(c.ClientCacheFilePath, readFileName)
	cached_file, err := os.Open(cacheFilePath)
	if err != nil {
		fmt.Println("ClientReadRequest ERROR:", err)
	}
	defer cached_file.Close()
	s := bufio.NewScanner(cached_file)

	numLinesRead := 0
	for s.Scan() {
		output += s.Text() + "\n"
		numLinesRead++
		if numLinesRead >= numLines {
			break
		}

	}

	// Either the end of file or an error. Break.
	if err != nil {
		fmt.Println("readFile ERROR:", err)
		return READ_ERROR
	}

	return output
}
