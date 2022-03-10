package nodecomm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

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

func createFile(path string, filename string, filetype string) {
	newfile, err := os.Create(path + "/" + filename + filetype)
	initLockContent(*newfile)
	defer newfile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func InitLockFiles(lock_path string, data_path string) {
	files, err := ioutil.ReadDir(data_path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			createFile(lock_path, file.Name(), ".lock")
		}
	}
}

func sequencerGenerator() string {
	return "sequence"
}

func AcquireWriteLock(filename string, lockPath string, lockdelay int, client_id int, sequencer bool) {
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
		// generate sequence and return
		l.Sequence = sequencerGenerator()
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

func ReleaseWriteLock(filename string, lockPath string, client_id int) {
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
