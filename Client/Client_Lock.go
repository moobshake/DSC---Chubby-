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
	timestamp time.Time
	lockDelay int
}

func (c *Client) isLockExpire(filename string) bool {
	timeDiff := time.Now().Sub(c.Locks[filename].timestamp).Seconds()
	if int(timeDiff) >= c.Locks[filename].lockDelay {
		fmt.Printf("%s lock expired for Client\n", c.Locks[filename].lockType)
		c.RelLock(filename)
		return true
	}
	return false
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
func (c *Client) RecvLock(sequencer string, lType string, timestamp string, lockdelay int) {
	var newLock lock

	if sequencer == "NotAvail" {
		fmt.Println("Lock was not available")
		return
	}

	ts := convertTime(timestamp)

	if lType == READ_CLI {
		newLock = lock{
			lockType:  READ_CLI,
			sequencer: sequencer,
			timestamp: ts,
			lockDelay: lockdelay,
		}
	} else if lType == WRITE_CLI {
		newLock = lock{
			lockType:  WRITE_CLI,
			sequencer: sequencer,
			timestamp: ts,
			lockDelay: lockdelay,
		}
	}

	fileName := strings.Split(sequencer, ",")[0]
	c.Locks[fileName] = newLock
	c.ListLocks()
}

// RelLock release Read Lock (Not implemented yet)
func (c *Client) RelLock(filename string) {

	// Release lock server

	// Delete from Client
	delete(c.Locks, filename)
}

// convert string to time
func convertTime(s string) time.Time {
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	t, err := time.Parse(layout, strings.Split(s, " m=")[0])
	if err != nil {
		fmt.Println(err)
	}
	return t
}
