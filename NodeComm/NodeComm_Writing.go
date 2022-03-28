package nodecomm

import (
	"fmt"
	"os"
	"path/filepath"

	pc "assignment1/main/protocchubby"
)

// TODO: Locking when available
func (n *Node) validateWriteLock() bool {
	return true
}

// This function appends the file information to the end of the file.
// If truncateFile is true, the file is truncated to 0 first before appending the information.
func (n *Node) writeToLocalFile(CliMsg *pc.ClientMessage, truncateFile bool) {
	fmt.Println("> server writing to file:", CliMsg.StringMessages, "for client", CliMsg.ClientID)
	// Create cache directory if not exists
	filePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)

	// Open file
	file, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("writeToLocalFile ERROR:", err)
	}
	defer file.Close()

	// truncate if needed
	if truncateFile {
		err := os.Truncate(filePath, 0)
		if err != nil {
			fmt.Println("writeToLocalFile TRUNCATE ERROR:", err)
		}
	}

	// write content
	if _, err := file.Write(CliMsg.FileBody.FileContent); err != nil {
		fmt.Println(err)
	}
}
