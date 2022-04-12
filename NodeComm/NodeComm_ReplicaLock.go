package nodecomm

import (
	pc "assignment1/main/protocchubby"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

func (n *Node) handleLockfromMaster(serverMsg *pc.ServerMessage) *pc.ServerMessage {
	filename := strings.Split(serverMsg.Lock.Sequencer, ",")[0]
	fmt.Printf("Replica received lock requested by %s\n", serverMsg.StringMessages)

	// Read from lock files
	file, err := ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Creating", filename, "file and lock")
		createFile(n.nodeDataPath, filename, "")
		createFile(n.nodeLockPath, filename, ".lock")
		file, _ = ioutil.ReadFile(n.nodeLockPath + "/" + filename + ".lock")
	}

	l := Lock{}

	err = json.Unmarshal([]byte(file), &l)
	if err != nil {
		fmt.Println(err)
	}

	// Append to lock file
	if err != nil {
		fmt.Println(err)
	}
	time_str := convertTime(serverMsg.Lock.TimeStamp)
	l_val := LockValues{Sequence: serverMsg.Lock.Sequencer, Timestamp: time_str, Lockdelay: int(serverMsg.Lock.LockDelay)}

	clientID, err := strconv.Atoi(serverMsg.StringMessages)
	if err != nil {
		fmt.Println(err)
	}

	switch serverMsg.Lock.Type {
	case pc.LockMessage_ReadLock:
		l.Read[clientID] = l_val
	case pc.LockMessage_WriteLock:
		l.Write[clientID] = l_val

	}

	data, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(n.nodeLockPath+"/"+filename+".lock", data, 0644)
	if err != nil {
		fmt.Println(err)
	}

	return &pc.ServerMessage{Type: pc.ServerMessage_Ack}
}

// Release lock
func (n *Node) relLockfromMaster(serverMsg *pc.ServerMessage) *pc.ServerMessage {

	clientId, _ := strconv.Atoi(serverMsg.StringMessages)
	lType := serverMsg.Lock.Type
	lSequencer := serverMsg.Lock.Sequencer
	result := n.ReleaseLockChecker(clientId, lType, lSequencer)

	if result == "error" {
		return &pc.ServerMessage{Type: pc.ServerMessage_Error}
	} else {
		return &pc.ServerMessage{Type: pc.ServerMessage_Ack}
	}
}

// convert string to time
func convertTime(s string) time.Time {
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	t, err := time.Parse(layout, strings.Split(s, " m=")[0])
	if err != nil {
		log.Fatalln(err)
	}
	return t
}
