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
	REL_LOCK       = "releaseLock"
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
			//Note from JW: This is not a clean exit. Some stray goroutines survive.
			break Main
		case WRITE_CLI:
			// Expect the file name to follow the read request token
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_FileWrite, StringMessages: tokenised[1]})
			}
		case READ_CLI:
			// Expect the file name to follow the read request token
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_FileRead, StringMessages: tokenised[1]})
			}
		case REQ_LOCK:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				if tokenised[1] == WRITE_CLI {
					c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_WriteLock, StringMessages: tokenised[2]})
				} else {
					c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_ReadLock, StringMessages: tokenised[2]})
				}
			}
		case REL_LOCK:
			if len(tokenised) < 2 {
				fmt.Println("Invalid Use of Command. Requires File Name Input")
			} else {
				c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_ReleaseLock, StringMessages: tokenised[1]})
			}
		case SUB:
			// subscriptions can be done without the client listener
			c.subscribe(tokenised[1:])
		case LIST_FILE_CLI:
			c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_ListFile})
		case LIST_LOCKS_CLI:
			c.DispatchControlClientMessage(&pc.ClientMessage{Type: pc.ClientMessage_ListLocks})

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
		fmt.Printf("'%s SUB_TYPE':\t Sends a subscription request. Type help %s for more info.\n", SUB, SUB)
		fmt.Printf("'%s':\t Lists the files available for the client\n", LIST_FILE_CLI)
		fmt.Printf("'%s':\t Lists the locks available for the client\n", LIST_LOCKS_CLI)
		fmt.Printf("'%s FILE_NAME':\t Release lock for FILE_NAME\n", REL_LOCK)
		fmt.Printf("'%s LOCK_TYPE FILE_NAME':\t Requests lock with LOCK_TYPE (%s/%s)for FILE_NAME\n", REQ_LOCK, READ_CLI, WRITE_CLI)
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
			StringMessages: args[1],
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
			StringMessages: args[1],
			ClientAddress:  &pc.PeerRecord{Address: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		}
		fmt.Printf("Client %d creating Subsciption Request %s \n", c.ClientID, "LockConflict")
	}
	res := c.DispatchClientMessage(c.MasterAdd, &cm)

	// Only do handle the response if there are no errors
	if res != nil {
		fmt.Printf("Master replied: %d, Message: %d, %s\n", res.Type, res.Message, res.StringMessages)

		if res.Type == pc.ClientMessage_RedirectToCoordinator {
			c.HandleMasterRediction(res)

		}
	} else {
		// Try to find a new master
		c.FindMaster()
		c.subscribe(args)
	}
}

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
			Spare:          int32(120),
		}
	case LIST_FILE_CLI:
		cm = pc.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     pc.ClientMessage_ListFile,
		}
		fmt.Printf("Client %d creating List File Request %s \n", c.ClientID, reqType)

	case REL_LOCK:
		// check if the lock exists
		filename := additionalArgs[0]
		if l, ok := c.Locks[filename]; ok {
			fmt.Printf("Lock for %s found! Sending request to release lock", filename)

			lt := pc.LockMessage_Empty
			switch l.lockType {
			case READ_CLI:
				lt = pc.LockMessage_ReadLock
			case WRITE_CLI:
				lt = pc.LockMessage_WriteLock
			}

			cm = pc.ClientMessage{
				ClientID: int32(c.ClientID),
				Type:     pc.ClientMessage_ReleaseLock,
				Lock:     &pc.LockMessage{Type: lt, Sequencer: l.sequencer, TimeStamp: l.timestamp.String(), LockDelay: int32(l.lockDelay)},
			}
		} else {
			fmt.Printf("Lock does not exist for %s\n", filename)
			return
		}

	}

	res := c.DispatchClientMessage(c.MasterAdd, &cm)

	// Only do handle the response if there are no errors
	if res != nil {
		if res.Type == pc.ClientMessage_ReadLock {
			c.RecvLock(res.Lock.Sequencer, "read", res.Lock.TimeStamp, int(res.Lock.LockDelay))
		} else if res.Type == pc.ClientMessage_WriteLock {
			c.RecvLock(res.StringMessages, "write", res.Lock.TimeStamp, int(res.Lock.LockDelay))
		} else if res.Type == pc.ClientMessage_ReleaseLock {
			fmt.Println("Releasing lock")
			c.RelLock(res.StringMessages) // pass back only filename to release lock
			fmt.Printf("Lock for %s is released\n", res.StringMessages)
		}

		fmt.Printf("Master replied: %d, Message: %d, %s\n", res.Type, res.Message, res.StringMessages)

		if res.Type == pc.ClientMessage_RedirectToCoordinator {
			c.HandleMasterRediction(res)
			c.ClientRequest(reqType, additionalArgs[0], additionalArgs[1])
		}
	} else {
		// Try to find a new master
		c.FindMaster()
		c.ClientRequest(reqType, additionalArgs[0], additionalArgs[1])
	}
}

func (c Client) ClientReadRequest(readFileName string) {
	// Check if the file is valid in cache
	// Darryl: also check if readLock has expired, if it is, send another request
	if c.ClientCacheValidation[readFileName] && c.isLockExpire(readFileName) == 2 {
		fmt.Println("> File", readFileName, "already exists in cache and is valid.")
	} else {
		// otherwise, request from master
		c.DispatchReadRequest(readFileName)
	}
}
