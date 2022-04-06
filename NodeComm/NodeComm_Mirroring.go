package nodecomm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	pc "assignment1/main/protocchubby"
)

const configLockFolderPath = "locks/"
const configDataFolderPath = "data/"

func (n *Node) MirrorDispatcher(msg *pc.CoordinationMessage) {
	go n.MirrorSource(msg.FromPRecord)
}

func (n *Node) MirrorSource(dPRecord *pc.PeerRecord) {
	lockFiles, err := os.Stat(configLockFolderPath)
	if err != nil {
		fmt.Println("Unable to find or access locks folder. ", err)
		return
	}
	dataFiles, err := os.Stat(configDataFolderPath)
	if err != nil {
		fmt.Println("Unable to find or access data folder. ", err)
		return
	}
	n.DispatchFolderRecords(dPRecord, configLockFolderPath, lockFiles)
	n.DispatchFolderRecords(dPRecord, configDataFolderPath, dataFiles)
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
			fPath := folderPath + f.Name() + "/"
			n.DispatchFolderRecords(dPRec, fPath, f)
		} else {
			fileName := f.Name()
			filePath := folderPath + fileName
			checksum := getFileChecksum(filePath)
			nMirrorRecord := &pc.MirrorRecord{FilePath: filePath, CheckSum: checksum}
			nMirrorRecords = append(nMirrorRecords, nMirrorRecord)
		}
	}
	n.DispatchCoordinationMessage(dPRec, &pc.CoordinationMessage{Type: pc.CoordinationMessage_MirrorRecord, MirrorRecords: nMirrorRecords})
}

func (n *Node) MirrorSink(MRecs []*pc.MirrorRecord) {
	if len(MRecs) == 0 {
		return
	}
	for _, MRecord := range MRecs {
		_, err := os.Stat(MRecord.FilePath)
		if err != nil || os.IsNotExist(err) {
			//File does not exist
			MRecs := []*pc.MirrorRecord{MRecord}
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
		_, err = ioutil.ReadFile(MRecord.FilePath)
		if err != nil {
			//Error reading
			MRecs := []*pc.MirrorRecord{MRecord}
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
		checksum := getFileChecksum(MRecord.FilePath)
		if !checkFileSumSame(MRecord.CheckSum, checksum) {
			//File not same
			MRecs := []*pc.MirrorRecord{MRecord}
			n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqFile, MirrorRecords: MRecs})
			continue
		}
	}
}
