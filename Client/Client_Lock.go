package client

import (
	"fmt"
	"strings"
	"time"
)

// Locks tracking
type lock struct {
	lockType  string
	sequencer string
}

func (c *Client) LockChecker() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			c.checker()
		}
	}
}

func (c *Client) checker() {
	for i := range c.Locks {
		fmt.Println(i)
	}
}

// ListLocks lists the current locks that client is holding
func (c *Client) ListLocks() {
	if len(c.Locks) == 0 {
		fmt.Printf("> Client is currently holding no locks\n")
	} else {
		fmt.Printf("> Client is currently holding:\n")
		for i := range c.Locks {
			fmt.Printf("%v\n", c.Locks[i])
		}
	}
}

// RecvLock receives Locks
func (c *Client) RecvLock(sequencer string, lType string) {
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
	c.Locks[fileName] = newLock
	c.ListLocks()
}

// RelLock release Read Lock (Not implemented yet)
func (C *Client) RelLock(filename string) {

	// Release lock server

	// Delete from Client
	delete(C.Locks, filename)
}
