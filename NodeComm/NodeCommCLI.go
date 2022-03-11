package nodecomm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
			n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_StopListening})
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
		fmt.Println("'exit':\t Exit program.")
		fmt.Println("'online':\t Onlines the node.")
		fmt.Println("'offline':\t Offlines the node.")
		fmt.Println("'config':\t Configuring node parameters.")
		fmt.Println("'addPeer':\t Adds a record about a peer node.")
		fmt.Println("'delPeer':\t Deletes a peer node record.")
		fmt.Println("'getPeer':\t Get the list of peer records.")
		fmt.Println("'msg':\t\t Send a plain text message to a peer node.")
		fmt.Println("'getStatus':\t Obtain information about the node.")
		fmt.Println("'startElection':\t Used to manually trigger an election.")
		fmt.Println("'help':\t Prints this menu.")
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
	}
}

func (n *Node) online() {
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_JoinNetwork})
}

func (n *Node) offline() {
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_LeaveNetwork})
}

func (n *Node) addPeer(params []string) {
	var nPRecs []*PeerRecord
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
		nPRec := PeerRecord{
			Id:      int32(peerID),
			Address: peerAddr,
			Port:    peerPort,
		}
		nPRecs = append(nPRecs, &nPRec)
	}
	response := n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_AddPeer, ParamsBody: &ParamsBody{PeerRecords: nPRecs}})
	if response.Type == ControlMessage_Okay {
		fmt.Println("Done")
	}
}

func (n *Node) delPeer(params []string) {
	var nPRecs []*PeerRecord
	for i, param := range params {
		if i == 0 {
			continue
		}
		peerID, _ := strconv.Atoi(param)
		nPRec := PeerRecord{Id: int32(peerID)}
		nPRecs = append(nPRecs, &nPRec)
	}
	response := n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_DelPeer, ParamsBody: &ParamsBody{PeerRecords: nPRecs}})
	if response.Type == ControlMessage_Okay {
		fmt.Println("Done")
	}
}

func (n *Node) getPeers() {
	response := n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_GetParams})
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
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_Message, Spare: int32(id), Comment: nMsg})
}

func (n *Node) getStatus() {
	response := n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_GetStatus})
	fmt.Println("id: " + strconv.Itoa(int(response.ParamsBody.MyPRecord.Id)) + "\tidOfMaster: " + strconv.Itoa(int(response.ParamsBody.IdOfMaster)))
}

func (n *Node) configParams(params []string) {
	response := n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_GetParams})
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
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_UpdateParams, ParamsBody: pBody})
}

func (n *Node) startElection() {
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_StartElection})
}
