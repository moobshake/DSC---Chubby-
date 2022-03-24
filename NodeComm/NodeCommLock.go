package nodecomm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	cp "github.com/otiai10/copy"
)

type LockValues struct {
	sequence  string
	lockdelay int
}

type Lock struct {
	Write map[int]LockValues // only 1 client can hold
	Read  map[int]LockValues // multiple client can hold
}

// create lock directory
func InitDirectory(path string, data bool) {
	_, err := os.Stat(path)
	if err != nil {
		e := os.Mkdir(path, 0755)
		if e != nil {
			log.Fatal(e)
		}
	}
	if data {
		err = cp.Copy("sample_data", path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// base on the file names, create a lock JSON with content
func initLockContent(file os.File) {
	l := Lock{Write: make(map[int]LockValues), Read: make(map[int]LockValues)}
	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(file.Name(), data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// helper function to create files
func createFile(path string, filename string, filetype string) {
	newfile, err := os.Create(path + "/" + filename + filetype)
	initLockContent(*newfile)
	defer newfile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// initialise lock files
func InitLockFiles(lock_path string, data_path string) {
	files, err := ioutil.ReadDir(data_path + "/")
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	f := ""
	for _, file := range files {
		f += file.Name() + "\n"
	}
	return f
}

// filename, mode (exclusive/shared), lock generation number
func sequencerGenerator(filename string, mode string, lock_gen_num int) string {
	return filename + ":" + mode + ":" + strconv.Itoa(lock_gen_num)
}

// acquire write lock
func (n *Node) AcquireWriteLock(filename string, client_id int, lockdelay int) (bool, string) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		log.Fatal(err)
	}

	// if write lock is currently held
	if len(l.Write) != 0 {
		return false, ""
	}

	// max lock delay of 10 seconds
	if lockdelay > 10 || lockdelay < 0 { // 10 second upper bound
		return false, ""
	}

	n.lockGenerationNumber++
	s := sequencerGenerator(filename, "exclusive", n.lockGenerationNumber)

	l.Write[client_id] = LockValues{sequence: s, lockdelay: lockdelay}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return true, s
}

// release the write lock
func (n *Node) ReleaseWriteLock(filename string, client_id int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		log.Fatal(err)
	}

	delete(l.Write, client_id)

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// acquire read lock
func (n *Node) AcquireReadLock(filename string, client_id int, lockdelay int) (bool, string) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		log.Fatal(err)
	}

	n.lockGenerationNumber++

	s := sequencerGenerator(filename, "shared", n.lockGenerationNumber)

	l.Read[client_id] = LockValues{sequence: s, lockdelay: lockdelay}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return true, s
}

// release read lock
func (n *Node) ReleaseReadLock(filename string, client_id int) {
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}
	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		log.Fatal(err)
	}

	delete(l.Read, client_id)

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
