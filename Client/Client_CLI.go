package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	pc "assignment1/main/protocchubby"
)

const (
	// CLI keywords
	EXIT_CLI       = "exit"
	READ_CLI       = "read"
	WRITE_CLI      = "write"
	REQ_LOCK       = "requestLock"
	SUB            = "sub"
	LIST_FILE_CLI  = "ls"
	LIST_LOCKS_CLI = "ll"
	TRUE_CLI       = "true"
	FALSE_CLI      = "false"
)

func (c *Client) startCLI() {
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
			// Expect the file name to follow the read request token
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				// by default modify the file
				modifyFile := true

				if len(tokenised) >= 3 {
					modifyFile = tokenised[2] == TRUE_CLI
				}

				c.sendClientWriteRequest(tokenised[1], modifyFile)
			}
		case READ_CLI:
			// Expect the file name to follow the read request token
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				c.DispatchReadRequest(tokenised[1])
			}
		case REQ_LOCK:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				c.ClientRequest(REQ_LOCK, tokenised[1], tokenised[2])
			}
		case SUB:
			c.subscribe(tokenised[1:])
		case LIST_FILE_CLI:
			c.ClientRequest(LIST_FILE_CLI)
		case LIST_LOCKS_CLI:
			c.ListLocks()
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
		fmt.Printf("'%s':\t Exit program.\n", EXIT_CLI)
		fmt.Printf("'%s FILE_NAME [modify file? %s/%s]':\t Client sends Write Request to Master.\n", WRITE_CLI, TRUE_CLI, FALSE_CLI)
		fmt.Printf("'%s FILE_NAME':\t Client sends Read Request to Master.\n", READ_CLI)
		return
	}
	switch params[1] {
	case SUB:
		fmt.Println("Sends a subscription request.")
		fmt.Println("\tSyntax: sub <type> <args>")
		fmt.Println("\tSubscription Types:")
		fmt.Println("\t\t'FileMod [file]':\t Subscribe to file modifications of a file.")
		fmt.Println("\t\t'LockAcquire [lock]':\t Subscribe to lock aquisitions of a lock.")
		fmt.Println("\t\t'LockConflict [lock]':\t Subscribe to lock conflicts of a lock.")
		fmt.Println("\t\t'MasterFailover':\t Subscribe to masterfailover events.")
	}
}

func (c Client) subscribe(args []string) {
	if len(args) < 1 {
		fmt.Println("No arguments were provided.")
		return
	}
	var cm pc.ClientMessage
	switch args[0] {
	case "MasterFailover":
		cm = pc.ClientMessage{
			ClientID:       int32(c.ClientID),
			Type:           pc.ClientMessage_SubscribeMasterFailover,
			StringMessages: "",
			ClientAddress:  &pc.PeerRecord{Address: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		}
		fmt.Printf("Client %d creating Subsciption Request %s \n", c.ClientID, "MasterFailover")
	case "FileMod":
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires File Name Input")
			return
		}
		cm = pc.ClientMessage{
			ClientID:       int32(c.ClientID),
			Type:           pc.ClientMessage_SubscribeFileModification,
			StringMessages: args[1],
			ClientAddress:  &pc.PeerRecord{Address: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		}
		fmt.Printf("Client %d creating Subsciption Request %s \n", c.ClientID, "FileMod")
	case "LockAcquire":
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires File Name Input")
			return
		}
		cm = pc.ClientMessage{
			ClientID:       int32(c.ClientID),
			Type:           pc.ClientMessage_SubscribeLockAquisition,
			StringMessages: args[0],
			ClientAddress:  &pc.PeerRecord{Address: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		}
		fmt.Printf("Client %d creating Subsciption Request %s \n", c.ClientID, "LockAcquire")
	case "LockConflict":
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires File Name Input")
			return
		}
		cm = pc.ClientMessage{
			ClientID:       int32(c.ClientID),
			Type:           pc.ClientMessage_SubscribeLockConflict,
			StringMessages: args[0],
			ClientAddress:  &pc.PeerRecord{Address: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		}
		fmt.Printf("Client %d creating Subsciption Request %s \n", c.ClientID, "LockConflict")
	}
	res := c.DispatchClientMessage(c.MasterAdd, &cm)

	if res.Type == 114 {
		c.RecvLock(res.StringMessages, "read")
	} else if res.Type == 115 {
		c.RecvLock(res.StringMessages, "write")
	}

	fmt.Printf("Master replied: %d, Message: %d, %s\n", res.Type, res.Message, res.StringMessages)
}

//Message from Jia Wei: This function is a bit weird/inefficient because either
//the way this function is implemented or the way the client CLI is implemented
//does not make sense.
//In the client CLI above, there is already a switch case that picks the correct
//branch based on user input. Thus, it makes sense to have the branch directly
//does what the user input wants (such as directly dispatching the message)
//It does not make sense to switch case the CLI and then to switch case again
//in the ClientRequest.

// ClientRequest - Making request
// Types - Subsciptions, Lock Requests
// Read is not processed here - go to ClientRead.go
// Write is not processed here - go to ClientWrite.go
// 1 input for AdditionalArgs is needed for file and lock subscriptions
func (c Client) ClientRequest(reqType string, additionalArgs ...string) {

	var cm pc.ClientMessage

	switch reqType {
	case REQ_LOCK:

		// Expected arguments
		// input: REQ_LOCK READ_CLI file_name

		var lockType pc.ClientMessage_MessageType

		if additionalArgs[0] == READ_CLI {
			lockType = pc.ClientMessage_ReadLock
		} else if additionalArgs[0] == WRITE_CLI {
			lockType = pc.ClientMessage_WriteLock
		}

		cm = pc.ClientMessage{
			ClientID:       int32(c.ClientID),
			Type:           lockType,
			StringMessages: additionalArgs[1],
		}
	case LIST_FILE_CLI:
		cm = pc.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     pc.ClientMessage_ListFile,
		}
		fmt.Printf("Client %d creating List File Request %s \n", c.ClientID, reqType)
	}

	res := c.DispatchClientMessage(c.MasterAdd, &cm)

	if res.Type == 114 {
		c.RecvLock(res.StringMessages, "read")
	} else if res.Type == 115 {
		c.RecvLock(res.StringMessages, "write")
	}

	fmt.Printf("Master replied: %d, Message: %d, %s\n", res.Type, res.Message, res.StringMessages)
}
