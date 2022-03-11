package main

import (
	Client "assignment1/main/Client"
	NodeComm "assignment1/main/NodeComm"
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	// server := flag.Bool("server", false, "Set to true to be Server.")
	// flag.Parse()

	// beServerFlag := *server
	// testSkeletalConn(beServerFlag)

	// CLI Reader
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Choose type to activate (c/s)")
	fmt.Println("c - Client Node")
	fmt.Println("s - Server Node")
	var char string
	fmt.Print("> ")

Loop:
	for scanner.Scan() {
		char = ""
		char = scanner.Text()
		tokenised := strings.Fields(char)

		if len(tokenised) == 0 {
			continue
		}

		switch tokenised[0] {
		case "c":
			fmt.Println("Entered Client")
			// Client Requirements
			cID := 1
			cIP := "127.0.0.1"
			cPort := "8000"
			c := Client.CreateClient(
				cID,
				cIP,
				cPort,
			)
			c.StartClient()

		case "s":
			fmt.Println("Entered Server")
			// Node Requirements
			// nodeIDFlag := flag.Int("id", -1, "Set the ID of the node.")
			// nodeIDOfMasterFlag := flag.Int("idOfMaster", -1, "Set the ID of the master node.")
			// nodeIPAddrFlag := flag.String("ipAddr", "127.0.0.1", "Set the IPv4 address of the node.")
			// nodePortFlag := flag.String("port", "9090", "Set the port of the node.")
			// verboseFlag := flag.Int("verbose", 1, "Set to 2 for verbose mode.")

			// flag.Parse()
			iFlag := 1
			mFlag := 1
			ipaFlag := "127.0.0.1"
			pFlag := "9090"
			vFlag := 1

			nodeIDFlag := &iFlag
			nodeIDOfMasterFlag := &mFlag
			nodeIPAddrFlag := &ipaFlag
			nodePortFlag := &pFlag
			verboseFlag := &vFlag

			// dataPath := "./data" + strconv.Itoa(*nodeIDFlag)
			// lockPath := "./lock" + strconv.Itoa(*nodeIDFlag)

			// NodeComm.InitDirectory(dataPath, true)
			// NodeComm.InitDirectory(lockPath, false)
			// NodeComm.InitLockFiles(lockPath, dataPath)
			// NodeComm.AcquireWriteLock("sample1.txt.lock", lockPath, 5, *nodeIDFlag, true)

			node := NodeComm.CreateNode(
				*nodeIDFlag,
				*nodeIDOfMasterFlag,
				*nodeIPAddrFlag,
				*nodePortFlag,
				*verboseFlag,
			)
			node.StartNode()
		case "exit":
			break Loop
		}
	}

}
