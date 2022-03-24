package client

import (
	"fmt"
	"strings"
)

// Received Locks
func recvLock(c Client, sequencer string, lType string) Client {
	var newLock lock

	if lType == "read" {
		fmt.Printf("%d received Read Lock\n", c.ClientID)
		newLock = lock{
			l_type:    "read",
			sequencer: sequencer,
		}
	} else if lType == "write" {
		fmt.Printf("%d received Write Lock\n", c.ClientID)
		newLock = lock{
			l_type:    "write",
			sequencer: sequencer,
		}
	}

	fileName := strings.Split(sequencer, ":")[0]
	c.Locks[fileName] = newLock

	return c
}

// Release Read Lock (Not implemented yet)
func relLock(c Client, filename string) {

	// Release lock server

	// Delete from Client
	delete(c.Locks, filename)
}
