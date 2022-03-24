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
	} else {
		fmt.Println("Lock was not available")
		return c
	}

	fileName := strings.Split(sequencer, ":")[0]
	c.Locks[fileName] = newLock
	fmt.Printf("Client %d lock status: %v", c.ClientID, c.Locks)
	return c
}

// Release Read Lock (Not implemented yet)
func relLock(c Client, filename string) {

	// Release lock server

	// Delete from Client
	delete(c.Locks, filename)
}
