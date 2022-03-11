package nodecomm

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

//Node is a logical structure for the bully node.
type Node struct {
	NodeCommServiceServer
	idOfMaster         int
	peerRecords        []*PeerRecord
	peerRecordsLock    sync.Mutex
	myPRecord          *PeerRecord
	isOnline           bool
	electionStatus     *ElectionStatus
	electionStatusLock sync.Mutex
	verbose            int
}

//CreateNode initialises a Node
func CreateNode(id, idOfMaster int, ipAddr, port string, verbose int) *Node {
	n := Node{
		idOfMaster:     idOfMaster,
		myPRecord:      &PeerRecord{Id: int32(id), Address: ipAddr, Port: port},
		electionStatus: &ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)},
		verbose:        verbose,
	}
	return &n
}

//StartNode starts the listener, initialises the params of the listener, and starts the UI.
func (n *Node) StartNode() {
	go n.startListener()
	pBody := ParamsBody{MyPRecord: n.myPRecord, IdOfMaster: int32(n.idOfMaster), ElectionStatus: n.electionStatus, Verbose: int32(n.verbose)}
	n.DispatchControlMessage(&ControlMessage{Type: ControlMessage_InitParams, ParamsBody: &pBody})
	time.Sleep(time.Second * 1)
	n.startCLI()
}

//startListener starts a *grpc.serve server as a listener.
func (n *Node) startListener() {
	if n.idOfMaster == -1 {
		fmt.Println("Unable to start without idOfMaster set!")
		return
	}
	fullAddress := n.myPRecord.Address + ":" + n.myPRecord.Port
	fmt.Println("Starting listener at ", fullAddress)
	fmt.Println("Master is node: ", n.idOfMaster)
	lis, err := net.Listen("tcp", fullAddress)
	if err != nil {
		log.Fatalf("Failed to hook into: %s. %v", fullAddress, err)
	}

	s := Node{}
	gServer := grpc.NewServer()
	RegisterNodeCommServiceServer(gServer, &s)

	if err := gServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}

//<><><><><><><><><><><><><><><><><><><><><>
//<><><> RPC Methods <><><>

//<> "Internal" Control API <>

//SendControlMessage: Channel for Control Messages
func (n *Node) SendControlMessage(ctx context.Context, cMsg *ControlMessage) (*ControlMessage, error) {
	if n.verbose == 2 {
		fmt.Println("Received ControlMessage:", cMsg)
	}
	switch cMsg.Type {
	case ControlMessage_GetStatus:
		pBody := ParamsBody{IdOfMaster: int32(n.idOfMaster), MyPRecord: n.myPRecord}
		return &ControlMessage{Type: ControlMessage_Okay, ParamsBody: &pBody}, nil
	case ControlMessage_GetPeers:
		pBody := ParamsBody{PeerRecords: n.peerRecords}
		return &ControlMessage{Type: ControlMessage_Okay, ParamsBody: &pBody}, nil
	case ControlMessage_InitParams:
		n.electionStatus = &ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)}
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_GetParams:
		pBody := ParamsBody{
			IdOfMaster: int32(n.idOfMaster),
			MyPRecord:  n.myPRecord,
			Verbose:    int32(n.verbose),
		}
		return &ControlMessage{Type: ControlMessage_Okay, ParamsBody: &pBody}, nil
	case ControlMessage_UpdateParams:
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_AddPeer:
		pRecords := cMsg.ParamsBody.PeerRecords
		n.mergePeerRecords(pRecords)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_DelPeer:
		pRecords := cMsg.ParamsBody.PeerRecords
		for _, pRec := range pRecords {
			n.deletePeerRecord(int(pRec.Id))
		}
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_Message:
		nMessage := NodeMessage{
			FromPRecord:  n.myPRecord,
			ToID:         cMsg.Spare,
			Type:         NodeMessage_Standard,
			StandardBody: &StandardBody{Message: cMsg.Comment}}
		n.DispatchMessage(n.getPeerRecord(int(cMsg.Spare), false), &nMessage)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_RawMessage:
		n.DispatchMessage(n.getPeerRecord(int(cMsg.NodeMessage.ToID), false), cMsg.NodeMessage)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_JoinNetwork:
		if !n.isOnline {
			n.onlineNode()
			return &ControlMessage{Type: ControlMessage_Okay}, nil
		} else {
			return &ControlMessage{Type: ControlMessage_Okay, Comment: "Node is already online!"}, nil
		}
	case ControlMessage_LeaveNetwork:
		if n.isOnline {
			n.offlineNode()
			return &ControlMessage{Type: ControlMessage_Okay}, nil
		} else {
			return &ControlMessage{Type: ControlMessage_Error, Comment: "Node is already offline!"}, nil
		}
	case ControlMessage_StartElection:
		n.spoofElection()
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	}

	return &ControlMessage{Type: ControlMessage_Error, Comment: "Unsupported function."}, nil
}

//Shutdown: Unimplemented as I found that a graceful shutdown is unimportant for this.
func (n *Node) Shutdown(ctx context.Context, cMsg *ControlMessage) (*ControlMessage, error) {
	if cMsg.Type != ControlMessage_StopListening {
		return &ControlMessage{Type: ControlMessage_Error, Comment: "Wrong channel."}, nil
	}
	return &ControlMessage{Type: ControlMessage_Okay}, nil
}

//<> "External" Peer-to-Peer API <>

//KeepAlive: Used to check if a node is alive or not.
func (n *Node) KeepAlive(ctx context.Context, inMsg *NodeMessage) (*NodeMessage, error) {
	if !n.isOnline {
		return &NodeMessage{Type: NodeMessage_Empty}, nil
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		return &NodeMessage{FromPRecord: n.myPRecord, Type: NodeMessage_NotInNetwork}, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case NodeMessage_KeepAlive:
		nMsg := NodeMessage{FromPRecord: n.myPRecord, Type: NodeMessage_KeepAlive, Comment: inMsg.Comment}
		return &nMsg, nil
	default:
		//Received a non-keepalive message on the keepalive channel
		return &NodeMessage{Type: NodeMessage_Error, Comment: "Unsupported function."}, nil
	}
}

//SendMessage: Channel for NodeMessages
func (n *Node) SendMessage(ctx context.Context, inMsg *NodeMessage) (*NodeMessage, error) {
	if !n.isOnline {
		return &NodeMessage{Type: NodeMessage_Empty}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Message:", inMsg)
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		nMsg := NodeMessage{FromPRecord: n.myPRecord, Type: NodeMessage_NotInNetwork}
		nMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
		nMsg.Spare = int32(n.idOfMaster)
		return &nMsg, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case NodeMessage_Standard:
		fmt.Println(inMsg.FromPRecord.Id, ": ", inMsg.StandardBody.Message)
		return &NodeMessage{FromPRecord: n.myPRecord, Type: NodeMessage_Ack}, nil
	default:
		//Received a non-standard message on the standard channel
		//Consider returning an error or doing something else
		return &NodeMessage{FromPRecord: n.myPRecord, Type: NodeMessage_Warning, Comment: "Message sent to wrong channel."}, nil
	}
}

//SendCoordinationMessage: Channgel for Coordination Messages
func (n *Node) SendCoordinationMessage(ctx context.Context, coMsg *CoordinationMessage) (*CoordinationMessage, error) {
	if !n.isOnline {
		return &CoordinationMessage{Type: CoordinationMessage_Empty}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Coordination Message:", coMsg)
	}
	nCoMsg := CoordinationMessage{FromPRecord: n.myPRecord, Type: CoordinationMessage_Error}
	//Node is NOT the master and it received an out of network message
	if !n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case CoordinationMessage_ReqToJoin: //Send redirection
			coordinatorRecord := []*PeerRecord{n.getPeerRecord(n.idOfMaster, false)}
			nCoMsg.Type = CoordinationMessage_RedirectToCoordinator
			nCoMsg.PeerRecords = coordinatorRecord
		default:
			nCoMsg.Type = CoordinationMessage_NotInNetwork
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is NOT the master and it received an in-network message
	} else if !n.IsMaster() {
		switch coMsg.Type {
		case CoordinationMessage_ApptNewCoordinator: //Accept if message was from master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.idOfMaster = int(coMsg.Spare)
				nCoMsg.Type = CoordinationMessage_Ack
				if n.IsMaster() {
					n.deletePeerRecord(int(coMsg.FromPRecord.Id))
					n.BroadcastPeerInformation()
				}
			} else {
				nCoMsg.Type = CoordinationMessage_NotMaster
				nCoMsg.Spare = int32(n.idOfMaster)
			}
		default:
			nCoMsg.Comment = "Unsupported function."
			nCoMsg.Type = CoordinationMessage_Error
		}
		//Node is master and it received an out of network message
	} else if n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case CoordinationMessage_ReqToJoin:
			pRec := n.getPeerRecord(int(coMsg.FromPRecord.Id), true)
			if pRec == nil { //Allow node to join, send peer records
				n.updatePeerRecords(coMsg.FromPRecord)
				nCoMsg.Type = CoordinationMessage_PeerInformation
				nCoMsg.PeerRecords = append(n.peerRecords, n.myPRecord)
				//Broadcast info to all nodes
				n.BroadcastPeerInformation()
			} else { //ID collision, REJECT node
				nCoMsg.Type = CoordinationMessage_RejectJoin
				nCoMsg.Comment = "ID Collision"
			}
		case CoordinationMessage_ReqToMerge:
			nCoMsg.Comment = "Unimplemented funcion."
		default:
			nCoMsg.Type = CoordinationMessage_NotInNetwork
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is master and it received an in-network message
	} else if n.IsMaster() {
		switch coMsg.Type {
		case CoordinationMessage_ReqToLeave:
			nCoMsg.Type = CoordinationMessage_Ack
			n.deletePeerRecord(int(coMsg.FromPRecord.Id))
			n.BroadcastPeerInformation()
		case CoordinationMessage_BadNodeReport:
			if n.badNodeHandler(n.getPeerRecord(int(coMsg.Spare), false)) {
				//Master node could reach allegedly downed node. Send peer information to
				//node that made the report in case it was a case of bad network details.
				go n.DispatchCoordinationMessage(coMsg.FromPRecord, &CoordinationMessage{Type: CoordinationMessage_PeerInformation, PeerRecords: append(n.peerRecords, n.myPRecord)})
			}
			nCoMsg.Type = CoordinationMessage_MessageType(CoordinationMessage_Ack)
		default:
			coMsg.Comment = "Unsupported operation."
		}
	}
	//Messages that apply for in-network nodes regardless of master status
	if n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case CoordinationMessage_PeerInformation: //Accept only if it came from the master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.overridePeerRecords(coMsg.PeerRecords)
			} else if n.IsMaster() { //Some other node think its the master, fix by re-triggering election
				n.spoofElection()
			} else { //The sendee node should not be sending a master-only message
				nCoMsg.Comment = "Sendee node is not coordinator."
				nCoMsg.Type = CoordinationMessage_NotMaster
			}
		case CoordinationMessage_ElectionResult, CoordinationMessage_ElectSelf, CoordinationMessage_RejectElect:
			n.electionHandler(coMsg)
		}
	}
	return &nCoMsg, nil
}

//SendClientMessage: Channel for ClientMessages
func (n *Node) SendClientMessage(ctx context.Context, CliMsg *ClientMessage) (*ClientMessage, error) {

	var ans int32

	// Replies with master address
	switch CliMsg.Type {
	case ClientMessage_FindMaster:
		// Find master
		ans = 5
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)
	case ClientMessage_FileRead:
		// Client request read
		ans = 6
		fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)
	case ClientMessage_FileWrite:
		ans = 7
		fmt.Printf("> Client %d requesting to write\n", CliMsg.ClientID)
	}
	return &ClientMessage{ClientID: CliMsg.ClientID, Type: ClientMessage_Ack, Message: int32(ans)}, nil
}
