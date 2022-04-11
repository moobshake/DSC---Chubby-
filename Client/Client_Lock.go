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
	softDelay int
	lockDelay int
}

func (c *Client) isLockExpire(filename string) int { //0:expired, 1:softExpired: 2:present
	if _, ok := c.Locks[filename]; !ok {
		// Client does not currently have the lock
		// It's probably been deleted by the LockChecker
		fmt.Println("Lock for", filename, "is either expired or it was never obtained.")
		return 0
	}
	//Lock is present, check if it's locally expired.
	timeDiff := time.Since(c.Locks[filename].timestamp).Seconds()
	if int(timeDiff) >= c.Locks[filename].softDelay {
		fmt.Println("This client does have the lock for", filename, "but it is locally expired.")
		return 1
	}
	return 2
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
	if len(c.Locks) == 0 {
		return
	}
	currentTime := time.Now()
	//find hard and soft expired locks
	var hardExpiredLocks []string
	var expiringLocks []string
	for lockName, lock := range c.Locks {
		hardExpireTime := lock.timestamp.Add(time.Duration(lock.lockDelay * 1000000000))
		hardExpireTimeLeft := hardExpireTime.Sub(currentTime)
		if hardExpireTimeLeft.Seconds() < 0 {
			hardExpiredLocks = append(hardExpiredLocks, lockName)
			continue
		}
		softExpireTime := lock.timestamp.Add(time.Duration(lock.softDelay * 1000000000))
		softExpireTimeLeft := softExpireTime.Sub(currentTime)
		if softExpireTimeLeft.Seconds() < 5 {
			expiringLocks = append(expiringLocks, lockName)
		}
	}
	//Clean up expired locks
	for _, expiredLock := range hardExpiredLocks {
		delete(c.Locks, expiredLock)
	}
	//Renew softly expiring locks if possible
	if c.DispatchClientKeepAlive() {
		for _, lockName := range expiringLocks {
			lock := c.Locks[lockName]
			lock.softDelay += lock.softDelay
			if lock.softDelay <= lock.lockDelay {
				c.Locks[lockName] = lock
			}
		}
	} else {
		fmt.Println("This client has", len(expiringLocks), "softly expiring locks that cannot be renewed because the primary server is uncontactable.")
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
			softDelay: 20,
			lockDelay: lockdelay,
		}
	} else if lType == WRITE_CLI {
		newLock = lock{
			lockType:  WRITE_CLI,
			sequencer: sequencer,
			timestamp: ts,
			softDelay: 20,
			lockDelay: lockdelay,
		}
	}

	fileName := strings.Split(sequencer, ",")[0]
	c.Locks[fileName] = newLock
	c.ListLocks()
}

// RelLock release Read Lock (Not implemented yet)
func (c *Client) RelLock(filename string) {
	// handle if filename is error
	if filename == "error" {
		fmt.Println("Tried to release lock, but sequencer is not correct or lock does not exists.")
		return
	}
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
