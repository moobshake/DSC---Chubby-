package nodecomm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	cp "github.com/otiai10/copy"
)

type LockValues struct {
	Sequence  string
	Timestamp time.Time
	Lockdelay int
}

type Lock struct {
	Write map[int]LockValues // only 1 client can hold
	Read  map[int]LockValues // multiple client can hold
}

// create lock directory
func (n *Node) InitDirectory(path string, data bool) {

	_, err := os.Stat(n.nodeRootPath)
	if err != nil {
		e := os.Mkdir(n.nodeRootPath, 0755)
		if e != nil {
			fmt.Println(e)
		}
	}
	_, err = os.Stat(path)
	if err != nil {
		e := os.Mkdir(path, 0755)
		if e != nil {
			fmt.Println(e)
		}
	}
	if data {
		err = cp.Copy("sample_data", path)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// base on the file names, create a lock JSON with content
func initLockContent(file os.File) {
	l := Lock{Write: make(map[int]LockValues), Read: make(map[int]LockValues)}
	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(file.Name(), data, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// helper function to create files
func createFile(path string, filename string, filetype string) {
	newfile, err := os.Create(path + "/" + filename + filetype)
	if filetype != "" {
		initLockContent(*newfile)
	}
	defer newfile.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// initialise lock files
func (n *Node) InitLockFiles(lock_path string, data_path string) {
	files, err := ioutil.ReadDir(data_path + "/")
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			createFile(lock_path, file.Name(), ".lock")
		}
	}
}

func list_files(data_path string) string {
	files, err := ioutil.ReadDir(data_path)
	if err != nil {
		fmt.Println(err)
	}
	f := ""
	for _, file := range files {
		f += file.Name() + "\n"
	}
	return f
}

// filename, mode (exclusive/shared), lock generation number
func sequencerGenerator(filename string, mode string, lock_gen_num int) string {
	return filename + "," + mode + "," + strconv.Itoa(lock_gen_num)
}

// acquire write lock
func (n *Node) AcquireWriteLock(filename string, client_id int, lockdelay int) (bool, string, string, int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
		// if no such lock exist, means no such file exist!
		fmt.Println("Creating", filename, "file and lock")
		createFile(n.nodeDataPath, filename, "")
		createFile(n.nodeLockPath, filename, ".lock")
		file, _ = ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
		return false, "", "", 0
	}

	// if write/read lock is currently held
	if len(l.Write) != 0 || len(l.Read) != 0 {
		fmt.Println("HELP")
		return false, "", "", 0
	}

	// max lock delay of 200 seconds
	if lockdelay > 200 || lockdelay < 0 { // 200 second upper bound
		fmt.Println("HELP ME")
		return false, "", "", 0
	}

	n.lockGenerationNumber++
	ts := time.Now()
	s := sequencerGenerator(filename, "exclusive", n.lockGenerationNumber)

	l.Write[client_id] = LockValues{Sequence: s, Lockdelay: lockdelay, Timestamp: ts}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		fmt.Println(err)
	}
	n.NoteClientIsAlive(int(client_id))
	return true, s, ts.String(), lockdelay
}

// release the write lock
func (n *Node) ReleaseWriteLock(filename string, client_id int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}

	delete(l.Write, client_id)

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// acquire read lock
func (n *Node) AcquireReadLock(filename string, client_id int, lockdelay int) (bool, string, string, int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
		return false, "", "", 0

	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
		return false, "", "", 0

	}

	// if write lock is current held, don't allow lock to be acquired
	if len(l.Write) != 0 {
		return false, "", "", 0
	}

	n.lockGenerationNumber++
	ts := time.Now()
	s := sequencerGenerator(filename, "shared", n.lockGenerationNumber)

	l.Read[client_id] = LockValues{Sequence: s, Lockdelay: lockdelay, Timestamp: ts}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		fmt.Println(err)
	}
	n.NoteClientIsAlive(int(client_id))
	return true, s, ts.String(), lockdelay
}

// release read lock
func (n *Node) ReleaseReadLock(filename string, client_id int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}

	delete(l.Read, client_id)

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
