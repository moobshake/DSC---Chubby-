package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const LOCK_PATH = "./lock"
const DATA_PATH = "./data"

type Lock struct {
	Write   int   // only 1 client can hold
	Read    []int // multiple client can hold
	Timeout int64 // timestamp for timeout
}

func initDirectory(path string) {
	_, err := os.Stat(path)
	if err != nil {
		e := os.Mkdir(path, 0755)
		if e != nil {
			log.Fatal(e)
		}
	}
}

func initLockContent(file os.File) {
	l := Lock{Write: -1, Read: []int{-1}, Timeout: 0}
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

func initLockFiles(lock_path string, data_path string) {
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

func main() {
	initDirectory(LOCK_PATH)
	initLockFiles(LOCK_PATH, DATA_PATH)
}
