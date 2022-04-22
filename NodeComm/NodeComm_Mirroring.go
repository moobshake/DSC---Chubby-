package nodecomm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	pc "assignment1/main/protocchubby"
)

func (n *Node) MirrorDispatcher(msg *pc.CoordinationMessage) {
	go n.MirrorSource(msg.FromPRecord)
}

func (n *Node) MirrorSource(dPRecord *pc.PeerRecord) {

	lockFiles, err := os.Stat(n.nodeLockPath)
	if err != nil {
		fmt.Println("Unable to find or access locks folder. ", err)
		return
	}
	dataFiles, err := os.Stat(n.nodeDataPath)
	if err != nil {
		fmt.Println("Unable to find or access data folder. ", err)
		return
	}
	n.DispatchFolderRecords(dPRecord, n.nodeRootPath, lockFiles)
	n.DispatchFolderRecords(dPRecord, n.nodeRootPath, dataFiles)
}

//SendFolderContents is a recursive function that dispatches mirror records for every file
//within a given directory recursively. Note: Empty folders will probably disappear.
func (n *Node) DispatchFolderRecords(dPRec *pc.PeerRecord, folderPath string, f fs.FileInfo) {
	filesInDir, err := ioutil.ReadDir(folderPath)

	if err != nil {
		fmt.Println("Unable to read file. ", err)
		return
	}
	if len(filesInDir) == 0 {
		return
	}
	var nMirrorRecords []*pc.MirrorRecord

	for _, file := range filesInDir {
		if file.IsDir() {
			fPath := filepath.Join(folderPath, f.Name())
			n.DispatchFolderRecords(dPRec, fPath, f)
			// Return otherwise, this recursion will also send out the dispatch message.
			return
		} else {
			fileName := file.Name()
			filePath := filepath.Join(folderPath, fileName)
			checksum := getFileChecksum(filePath)

			// Remove the first part of the directory as it is unique to the master
			var sharedFileName string

			if strings.Contains(filePath, "\\") {
				sharedFileNameSplit := strings.Split(filePath, "\\")[1:]
				sharedFileName = strings.Join(sharedFileNameSplit, "\\")
			} else {
				sharedFileNameSplit := strings.Split(filePath, "/")[1:]
				sharedFileName = strings.Join(sharedFileNameSplit, "/")
			}

			nMirrorRecord := &pc.MirrorRecord{FilePath: sharedFileName, CheckSum: checksum}
			nMirrorRecords = append(nMirrorRecords, nMirrorRecord)
		}
	}

	n.DispatchCoordinationMessage(dPRec, &pc.CoordinationMessage{Type: pc.CoordinationMessage_MirrorRecord, MirrorRecords: nMirrorRecords})
}

func (n *Node) MirrorSink(MRecs []*pc.MirrorRecord) {
	if len(MRecs) == 0 {
		return
	}
	fmt.Println("MirrorSink")

	for _, MRecord := range MRecs {
		// Get the correct file path for this machine os
		getCorrectFilePath(&MRecord.FilePath)

		replicaFilePath := filepath.Join(n.nodeRootPath, MRecord.FilePath)
		var filePath string
		if strings.Contains(MRecord.FilePath, "\\") {
			filePathSplit := strings.Split(MRecord.FilePath, "\\")
			filePath = strings.Join(filePathSplit, "\\")
		} else {
			filePath = MRecord.FilePath
		}

		// fmt.Println(replicaFilePath)

		_, err := os.Stat(replicaFilePath)
		if err != nil || os.IsNotExist(err) {
			//File does not exist
			MRecs := []*pc.MirrorRecord{MRecord}
			fmt.Println("MirrorSink: not exist")
			n.outstandingFiles[filePath] = MRecord
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
		_, err = ioutil.ReadFile(replicaFilePath)
		if err != nil {
			//Error reading
			MRecs := []*pc.MirrorRecord{MRecord}
			fmt.Println("MirrorSink: cannot access")
			n.outstandingFiles[filePath] = MRecord
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
		checksum := getFileChecksum(replicaFilePath)
		if !checkFileSumSame(MRecord.CheckSum, checksum) {
			fmt.Println("MirrorSink: invalid checksum")

			//File not same
			MRecs := []*pc.MirrorRecord{MRecord}
			n.outstandingFiles[filePath] = MRecord
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
	}
}

func (n *Node) MirrorService(mirrorInterval int) {
	for {
		time.Sleep(time.Second * time.Duration(mirrorInterval))
		if n.IsMaster() {
			continue
		}
		if !n.isOnline {
			break
		}
		n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqToMirror})
	}
}
