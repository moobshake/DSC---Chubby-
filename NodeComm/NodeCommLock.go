package nodecomm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	cp "github.com/otiai10/copy"
)

const LOCK_PATH = "./lock"
const DATA_PATH = "./data"

type Lock struct {
	Write     int    // only 1 client can hold
	Read      []int  // multiple client can hold
	LockDelay int64  // timestamp for timeout
	Sequence  string // opaque byte-string
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
	l := Lock{Write: -1, Read: []int{-1}, LockDelay: 0, Sequence: ""}
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
func (n *Node) AcquireWriteLock(filename string, lockPath string, lockdelay int, client_id int, sequencer bool) {
	file, err := ioutil.ReadFile(lockPath + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}

	err = json.Unmarshal([]byte(file), &l)

	if err != nil {
		log.Fatal(err)
	}

	if l.Write == -1 {
		l.Write = client_id
		if lockdelay < 10 { // 10 second upper bound
			l.LockDelay = int64(lockdelay)
		}
	} else {
		return // held by some other client
	}

	if sequencer {
		// add 1 to lock gen number
		n.lockGenerationNumber++
		// generate sequence and return
		l.Sequence = sequencerGenerator(filename, "exclusive", n.lockGenerationNumber)
	}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(lockPath+"/"+filename, data, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

// release the write lock
func (n *Node) ReleaseWriteLock(filename string, lockPath string, client_id int) {
	file, err := ioutil.ReadFile(lockPath + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}

	l := Lock{}

	err = json.Unmarshal([]byte(file), &l)

	if err != nil {
		log.Fatal(err)
	}

	if l.Write == client_id {
		l.Write = -1
		l.LockDelay = -1
		l.Sequence = ""
	}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(lockPath+"/"+filename, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
