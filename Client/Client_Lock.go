package client

import (
	"fmt"
	"strings"
)

// Locks tracking
type lock struct {
	lockType  string
	sequencer string
}

// ListLocks lists the current locks that client is holding
func (C *Client) ListLocks() {
	if len(C.Locks) == 0 {
		fmt.Printf("> Client is currently holding no locks\n")
	} else {
		fmt.Printf("> Client is currently holding:\n")
		for i := range C.Locks {
			fmt.Printf("%v\n", C.Locks[i])
		}
	}
}

// RecvLock receives Locks
func (C *Client) RecvLock(sequencer string, lType string) {
	var newLock lock

	if lType == READ_CLI {
		newLock = lock{
			lockType:  READ_CLI,
			sequencer: sequencer,
		}
	} else if lType == WRITE_CLI {
		newLock = lock{
			lockType:  WRITE_CLI,
			sequencer: sequencer,
		}
	} else {
		fmt.Println("Lock was not available")
		return
	}

	fileName := strings.Split(sequencer, ":")[0]
	C.Locks[fileName] = newLock
	C.ListLocks()
}

// RelLock release Read Lock (Not implemented yet)
func (C *Client) RelLock(filename string) {

	// Release lock server

	// Delete from Client
	delete(C.Locks, filename)
}
