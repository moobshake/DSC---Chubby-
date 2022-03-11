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

	// Original
	// server := flag.Bool("server", false, "Set to true to be Server.")
	// flag.Parse()

	// beServerFlag := *server
	// testSkeletalConn(beServerFlag)
	// Node Requirements
	// nodeIDFlag := flag.Int("id", -1, "Set the ID of the node.")
	// nodeIDOfMasterFlag := flag.Int("idOfMaster", -1, "Set the ID of the master node.")
	// nodeIPAddrFlag := flag.String("ipAddr", "127.0.0.1", "Set the IPv4 address of the node.")
	// nodePortFlag := flag.String("port", "9090", "Set the port of the node.")
	// verboseFlag := flag.Int("verbose", 1, "Set to 2 for verbose mode.")
	// dataPath := "./data" + strconv.Itoa(*nodeIDFlag)
	// lockPath := "./lock" + strconv.Itoa(*nodeIDFlag)

	// NodeComm.InitDirectory(dataPath, true)
	// NodeComm.InitDirectory(lockPath, false)
	// NodeComm.InitLockFiles(lockPath, dataPath)
	// NodeComm.AcquireWriteLock("sample1.txt.lock", lockPath, 5, *nodeIDFlag, true)

	// node := NodeComm.CreateNode(
	// 	*nodeIDFlag,
	// 	*nodeIDOfMasterFlag,
	// 	*nodeIPAddrFlag,
	// 	*nodePortFlag,
	// 	*verboseFlag,
	// )
	// node.StartNode()

	// Yu Hui Testing Version
	// CLI Reader
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Choose type to activate (c/s1/s2/s3)")
	fmt.Println("c - Client Node")
	fmt.Println("s1 - Server 1 Node")
	fmt.Println("s2 - Server 2 Node")
	fmt.Println("s3 - Server 3 Node")

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

		case "s1":
			fmt.Println("Entered Server 1")
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

			node := NodeComm.CreateNode(
				*nodeIDFlag,
				*nodeIDOfMasterFlag,
				*nodeIPAddrFlag,
				*nodePortFlag,
				*verboseFlag,
			)
			node.StartNode()

		case "s2":
			fmt.Println("Entered Server 2")

			iFlag := 2
			mFlag := 1
			ipaFlag := "127.0.0.1"
			pFlag := "9091"
			vFlag := 1

			nodeIDFlag := &iFlag
			nodeIDOfMasterFlag := &mFlag
			nodeIPAddrFlag := &ipaFlag
			nodePortFlag := &pFlag
			verboseFlag := &vFlag

			node := NodeComm.CreateNode(
				*nodeIDFlag,
				*nodeIDOfMasterFlag,
				*nodeIPAddrFlag,
				*nodePortFlag,
				*verboseFlag,
			)
			node.StartNode()

		case "s3":
			fmt.Println("Entered Server 3")

			iFlag := 3
			mFlag := 1
			ipaFlag := "127.0.0.1"
			pFlag := "9092"
			vFlag := 1

			nodeIDFlag := &iFlag
			nodeIDOfMasterFlag := &mFlag
			nodeIPAddrFlag := &ipaFlag
			nodePortFlag := &pFlag
			verboseFlag := &vFlag

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

		fmt.Print("> ")
	}

}
