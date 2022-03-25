package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// CLI keywords
	EXIT_CLI                = "exit"
	READ_CLI                = "read"
	WRITE_CLI               = "write"
	REQ_LOCK                = "requestLock"
	SUB_FILE_MOD_CLI        = "subFileMod"
	SUB_LOCK_AQUIS_CLI      = "subLockAquis"
	SUB_LOCK_CONFLICT_CLI   = "subLockConflict"
	SUB_MASTER_FAILOVER_CLI = "subMasterFailover"
	LIST_FILE_CLI           = "ls"
)

//Implement a CLI here that is built on top of the ActualClientAPI.go
func (C *Client) startCLI() {
	scanner := bufio.NewScanner(os.Stdin)
	var userInput string
	fmt.Print("> ")

Main:
	for scanner.Scan() {
		userInput = ""
		userInput = scanner.Text()
		tokenised := strings.Fields(userInput)

		if len(tokenised) == 0 {
			continue
		}

		switch tokenised[0] {
		case EXIT_CLI:
			break Main
		case WRITE_CLI:
			C.ClientRequest(WRITE_CLI, tokenised[1])
		case READ_CLI:
			// Expect the file name to follow the read request token
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				C.sendClientReadRequest(tokenised[1])
			}
		case REQ_LOCK:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				C.ClientRequest(REQ_LOCK, tokenised[1], tokenised[2])
			}
		case SUB_MASTER_FAILOVER_CLI:
			C.ClientRequest(SUB_MASTER_FAILOVER_CLI)
		case SUB_FILE_MOD_CLI:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				C.ClientRequest(SUB_FILE_MOD_CLI, tokenised[1])
			}
		case SUB_LOCK_AQUIS_CLI:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				C.ClientRequest(SUB_LOCK_AQUIS_CLI, tokenised[1])
			}
		case SUB_LOCK_CONFLICT_CLI:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				C.ClientRequest(SUB_LOCK_CONFLICT_CLI, tokenised[1])
			}
		case LIST_FILE_CLI:
			C.ClientRequest(LIST_FILE_CLI)
		case "help":
			printHelp(tokenised)
		default:
			fmt.Printf("Invalid Input: %s\n", userInput)
		}

		fmt.Print("> ")
	}
}

func printHelp(params []string) {
	if len(params) == 1 {
		fmt.Printf("'%s':\t Exit program.", EXIT_CLI)
		fmt.Printf("'%s':\t Client sends Write Request to Master.\n", WRITE_CLI)
		fmt.Printf("'%s':\t Client sends Read Request to Master.\n", READ_CLI)

		// Event Subsciptions
		fmt.Printf("'%s':\t Client sends Master Failover Subscription Request to Master.\n", SUB_MASTER_FAILOVER_CLI)
		fmt.Printf("'%s':\t Client sends File Modification Subscription Request to Master.\n", SUB_FILE_MOD_CLI)
		fmt.Printf("'%s':\t Client sends Lock Aquisition Subscription Request to Master.\n", SUB_LOCK_AQUIS_CLI)
		fmt.Printf("'%s':\t Client sends Lock Conflict Subscription Request to Master.\n", SUB_LOCK_CONFLICT_CLI)
	}
}
