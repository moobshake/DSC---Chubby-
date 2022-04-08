package nodecomm

import (
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
		timeDiff := time.Now().Sub(l.Read[i].Timestamp).Seconds()
		if int(timeDiff) >= l.Read[i].Lockdelay {
			fmt.Printf("Read lock expired for Client %d\n", i)
			fn := strings.ReplaceAll(filename, ".lock", "")
			n.ReleaseReadLock(fn, i)
		}
	}
}
