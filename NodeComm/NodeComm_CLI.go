package nodecomm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	pc "assignment1/main/protocchubby"
)

//startCLI starts the CLI (mainloop)
func (n *Node) startCLI() {
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
		case "exit":
			n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_StopListening})
			break Main
		case "online":
			n.online()
		case "offline":
			n.offline()
		case "config":
			n.configParams(tokenised)
		case "addPeer":
			n.addPeer(tokenised)
		case "delPeer":
			n.delPeer(tokenised)
		case "getPeers":
			n.getPeers()
		case "msg":
			n.msg(tokenised)
		case "getStatus":
			n.getStatus()
		case "startElection":
			n.startElection()
		case "wakeUpNode":
			n.wakeUpNode(tokenised)
		case "publish":
			n.publish(tokenised[1:])
		case "help":
			printHelp(tokenised)
		default:
			fmt.Println("Invalid input: ", userInput)
		}
		fmt.Print("> ")
	}
}

func printHelp(params []string) {
	if len(params) == 1 {
		fmt.Println("Available Commands:")
		fmt.Println("'exit':\t\t\t\tExit program.")
		fmt.Println("'online':\t\t\tOnlines the node.")
		fmt.Println("'offline':\t\t\tOfflines the node.")
		fmt.Println("'config':\t\t\tConfiguring node parameters.")
		fmt.Println("'addPeer':\t\t\tAdds a record about a peer node.")
		fmt.Println("'delPeer':\t\t\tDeletes a peer node record.")
		fmt.Println("'getPeer':\t\t\tGet the list of peer records.")
		fmt.Println("'msg':\t\t\t\tSend a plain text message to a peer node.")
		fmt.Println("'getStatus':\t\t\tObtain information about the node.")
		fmt.Println("'startElection':\t\tUsed to manually trigger an election.")
		fmt.Println("'publish':\t\t\tUsed to manually publish an update to subscribed clients.")
		fmt.Println("'wakeUpNode [ADDR:PORT]':\tUsed to manually wake up another replica.")
		fmt.Println("'help':\t\t\t\tPrints this menu.")
		return
	}
	switch params[1] {
	case "config":
		fmt.Println("Configures parameters about this node.")
		fmt.Println("Enter arguments as such: flag=VALUE")
		fmt.Println("id\t\t=\tID_OF_THIS_NODE")
		fmt.Println("idOfMaster\t=\tID_OF_MASTER")
		fmt.Println("verbose\t\t=\t1 or 2 for off and on respectively")
	case "addPeer":
		fmt.Println("Allows you to add records about other peer nodes.")
		fmt.Println("Enter individual nodes in the below format:")
		fmt.Println("\tNODE_ID:NODE_IPv4_ADDRESS:NODE_PORT")
		fmt.Println("\teg: addPeer 1:127.0.0.1:9090")
	case "delPeer":
		fmt.Println("Allows you to delete records about other peer nodes.")
		fmt.Println("Enter the IDs of the individual nodes to be deleted:")
		fmt.Println("\teg: delPeer 1")
	case "getPeer":
		fmt.Println("Obtains information about the peer nodes that are known to this node.")
	case "msg":
		fmt.Println("Sends a plain text message to a peer node.")
		fmt.Println("msg NODE_ID MESSAGE")
	case "getStatus":
		fmt.Println("Obtains information about this node.")
	case "startElection":
		fmt.Println("Spoofs this node into starting an election by faking a Election Message.")
	case "online":
		fmt.Println("Instructs the node to join or create a network.")
	case "offline":
		fmt.Println("Instructs the node to leave its network.")
	case "publish":
		fmt.Println("Manually publishes a subscription update.")
		fmt.Println("\tSyntax: publish <subscription_update_type> <args>")
		fmt.Println("\tSubscription Update Types:")
		fmt.Println("\t\t'FileMod [file]':\t Used to manually publish a file modification.")
		fmt.Println("\t\t'LockAcquire [lock]':\t Used to manually publish a lock aquisition.")
		fmt.Println("\t\t'LockConflict [lock]':\t Used to manually publish a lock conflict.")
		fmt.Println("\t\t'MasterFailover':\t Used to manually publish a master failover.")
	}
}

func (n *Node) online() {
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_JoinNetwork})
}

func (n *Node) offline() {
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_LeaveNetwork})
}

func (n *Node) addPeer(params []string) {
	var nPRecs []*pc.PeerRecord
	for i, param := range params {
		if i == 0 {
			continue
		}
		tokenised := strings.Split(param, ":")
		if len(tokenised) != 3 {
			continue
		}
		peerID, _ := strconv.Atoi(tokenised[0])
		peerAddr := tokenised[1]
		peerPort := tokenised[2]
		nPRec := pc.PeerRecord{
			Id:      int32(peerID),
			Address: peerAddr,
			Port:    peerPort,
		}
		nPRecs = append(nPRecs, &nPRec)
	}
	response := n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_AddPeer, ParamsBody: &pc.ParamsBody{PeerRecords: nPRecs}})
	if response.Type == pc.ControlMessage_Okay {
		fmt.Println("Done")
	}
}

func (n *Node) delPeer(params []string) {
	var nPRecs []*pc.PeerRecord
	for i, param := range params {
		if i == 0 {
			continue
		}
		peerID, _ := strconv.Atoi(param)
		nPRec := pc.PeerRecord{Id: int32(peerID)}
		nPRecs = append(nPRecs, &nPRec)
	}
	response := n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_DelPeer, ParamsBody: &pc.ParamsBody{PeerRecords: nPRecs}})
	if response.Type == pc.ControlMessage_Okay {
		fmt.Println("Done")
	}
}

func (n *Node) getPeers() {
	response := n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_GetPeers})
	fmt.Println(response.ParamsBody.PeerRecords)
}

func (n *Node) msg(params []string) {
	if len(params) < 3 {
		fmt.Println("Invalid message.")
		return
	}
	id, _ := strconv.Atoi(params[1])
	nMsg := ""
	for _, word := range params[2:] {
		nMsg = nMsg + word + " "
	}
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_Message, Spare: int32(id), Comment: nMsg})
}

func (n *Node) getStatus() {
	response := n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_GetStatus})
	fmt.Println("id: " + strconv.Itoa(int(response.ParamsBody.MyPRecord.Id)) + "\tidOfMaster: " + strconv.Itoa(int(response.ParamsBody.IdOfMaster)))
}

func (n *Node) configParams(params []string) {
	response := n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_GetParams})
	pBody := response.ParamsBody
	for i, param := range params {
		if i == 0 {
			continue
		}
		tokenised := strings.Split(param, "=")
		if len(tokenised) != 2 {
			continue
		}
		flag := tokenised[0]
		arg := tokenised[1]
		switch flag {
		case "id":
			nID, _ := strconv.Atoi(arg)
			pBody.MyPRecord.Id = int32(nID)
			n.myPRecord.Id = int32(nID) //Honestly this is unimportant because the client copy is not used
		case "idOfMaster":
			nID, _ := strconv.Atoi(arg)
			pBody.IdOfMaster = int32(nID)
			n.idOfMaster = nID //Honestly this is unimportant because the client copy is not used
		case "verbose":
			option, _ := strconv.Atoi(arg)
			pBody.Verbose = int32(option)
			n.verbose = option //Honestly this is unimportant because the client copy is not used
		}

	}
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_UpdateParams, ParamsBody: pBody})
}

func (n *Node) startElection() {
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_StartElection})
}

func (n *Node) wakeUpNode(params []string) {
	var nPRecs []*pc.PeerRecord
	for i, param := range params {
		if i == 0 {
			continue
		}
		tokenised := strings.Split(param, ":")
		if len(tokenised) != 2 {
			continue
		}
		peerAddr := tokenised[0]
		peerPort := tokenised[1]
		nPRec := pc.PeerRecord{
			Address: peerAddr,
			Port:    peerPort,
		}
		nPRecs = append(nPRecs, &nPRec)
	}

	if len(nPRecs) == 0 {
		fmt.Println("No valid peers to wake up.")
		return
	}

	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_WakeUpNode, ParamsBody: &pc.ParamsBody{PeerRecords: nPRecs}})
}

func (n *Node) publish(args []string) {
	if len(args) < 1 {
		fmt.Println("No arguments provided.")
		return
	}
	switch args[0] {
	case "FileMod":
		// Testing Events
		// Second arg should be the file name
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires File Name Input")
		} else {
			n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_PublishFileModification, Comment: args[1]})
		}
	case "LockAcquire":
		// Testing Events
		// Second arg should be the lock name
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires Lock Name Input")
		} else {
			n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_PublishLockAquisition, Comment: args[1]})
		}
	case "LockConflict":
		// Testing Events
		// Second arg should be the lock name
		if len(args) < 2 {
			fmt.Println("Invalid Use of Command. Requires Lock Name Input")
		} else {
			n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_PublishLockConflict, Comment: args[1]})
		}
	case "MasterFailover":
		// Testing Events
		n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_PublishMasterFailover})
	}
}
