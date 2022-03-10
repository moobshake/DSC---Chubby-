package main

import (
	"flag"
	"strconv"
	"sync"

	NodeComm "assignment1/main/NodeComm"
)

var wg sync.WaitGroup

func main() {
	// server := flag.Bool("server", false, "Set to true to be Server.")
	// flag.Parse()

	// beServerFlag := *server
	// testSkeletalConn(beServerFlag)

	nodeIDFlag := flag.Int("id", -1, "Set the ID of the node.")
	nodeIDOfMasterFlag := flag.Int("idOfMaster", -1, "Set the ID of the master node.")
	nodeIPAddrFlag := flag.String("ipAddr", "127.0.0.1", "Set the IPv4 address of the node.")
	nodePortFlag := flag.String("port", "9090", "Set the port of the node.")
	verboseFlag := flag.Int("verbose", 1, "Set to 2 for verbose mode.")

	flag.Parse()

	dataPath := "./data" + strconv.Itoa(*nodeIDFlag)
	lockPath := "./lock" + strconv.Itoa(*nodeIDFlag)

	NodeComm.InitDirectory(dataPath, true)
	NodeComm.InitDirectory(lockPath, false)
	NodeComm.InitLockFiles(lockPath, dataPath)

	node := NodeComm.CreateNode(
		*nodeIDFlag,
		*nodeIDOfMasterFlag,
		*nodeIPAddrFlag,
		*nodePortFlag,
		*verboseFlag,
	)

	node.StartNode()

}
