package client

import (
	"fmt"
	"strings"
)

// List the current locks that client is holding
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

// Received Locks
func (C *Client) RecvLock(sequencer string, lType string) {
	var newLock lock

	if lType == "read" {
		newLock = lock{
			l_type:    "read",
			sequencer: sequencer,
		}
	} else if lType == "write" {
		newLock = lock{
			l_type:    "write",
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

// Release Read Lock (Not implemented yet)
func (C *Client) RelLock(filename string) {

	// Release lock server

	// Delete from Client
	delete(C.Locks, filename)
}
