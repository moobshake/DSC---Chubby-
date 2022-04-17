package nodecomm

import (
	pc "assignment1/main/protocchubby"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// go routine to check for lock expiry
func (n *Node) LockChecker() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			n.checkFiles(n.nodeLockPath)
		}
	}
}

// check for files in the lock folder
func (n *Node) checkFiles(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		f, err := ioutil.ReadFile(path + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
		}
		l := Lock{}
		err = json.Unmarshal([]byte(f), &l)
		if err != nil {
			fmt.Println(err)
		}

		// check if write/read lock is available
		if len(l.Write) == 1 {
			n.checkWriteLock(l, file.Name())
		}
		if len(l.Read) != 0 {
			n.checkReadLock(l, file.Name())
		}
	}
}

// check if write lock expired
func (n *Node) checkWriteLock(l Lock, filename string) {
	for i := range l.Write {
		// check if client is still active
		// if not active just release lock
		if !n.isClientActiveLock(i) {
			fmt.Println("Client is not active just release write lock")
			fmt.Println(n._activeClients)
			fmt.Println(n.activeClients)
			fn := strings.ReplaceAll(filename, ".lock", "")
			n.ReleaseWriteLock(fn, i)
		}
		timeDiff := time.Now().Sub(l.Write[i].Timestamp).Seconds()
		if int(timeDiff) >= l.Write[i].Lockdelay {
			fmt.Printf("Write lock expired for Client %d\n", i)
			fn := strings.ReplaceAll(filename, ".lock", "")
			n.ReleaseWriteLock(fn, i)
		}
	}
}

// check if read locks expired
func (n *Node) checkReadLock(l Lock, filename string) {
	for i := range l.Read {
		// check if client is still active
		// if not active just release lock
		if !n.isClientActiveLock(i) {
			fmt.Println("Client is not active just release read lock")
			fn := strings.ReplaceAll(filename, ".lock", "")
			n.ReleaseReadLock(fn, i)
		}
		timeDiff := time.Now().Sub(l.Read[i].Timestamp).Seconds()
		if int(timeDiff) >= l.Read[i].Lockdelay {
			fmt.Printf("Read lock expired for Client %d\n", i)
			fn := strings.ReplaceAll(filename, ".lock", "")
			n.ReleaseReadLock(fn, i)
		}
	}
}

// Handles the releasing of locks
// Returns file name
func (n *Node) ReleaseLockChecker(clientId int, lType pc.LockMessage_LockType, lSequencer string) string {
	// Check if lock is valid
	filename := strings.Split(lSequencer, ",")[0]

	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}

	if lType == pc.LockMessage_ReadLock {
		if readlock, ok := l.Read[clientId]; ok {
			if readlock.Sequence == lSequencer {
				fmt.Printf("Received request to release read lock for %s for client %d\n", filename, clientId)
				n.ReleaseReadLock(filename, clientId)
				return filename
			}
		}
	} else if lType == pc.LockMessage_WriteLock {
		if writelock, ok := l.Write[clientId]; ok {
			if writelock.Sequence == lSequencer {
				fmt.Printf("Received request to release write lock for %s for client %d\n", filename, clientId)
				n.ReleaseWriteLock(filename, clientId)
				return filename
			}
		}
	}
	return "error"
}
