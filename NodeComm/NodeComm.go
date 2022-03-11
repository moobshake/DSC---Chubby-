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
	n.DispatchControlMessage(&ControlMessage{Type: int32(InitParams), ParamsBody: &pBody})
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
	case int32(GetStatus):
		pBody := ParamsBody{IdOfMaster: int32(n.idOfMaster), MyPRecord: n.myPRecord}
		return &ControlMessage{Type: int32(Okay), ParamsBody: &pBody}, nil
	case int32(GetPeers):
		pBody := ParamsBody{PeerRecords: n.peerRecords}
		return &ControlMessage{Type: int32(Okay), ParamsBody: &pBody}, nil
	case int32(InitParams):
		n.electionStatus = &ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)}
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(GetParams):
		pBody := ParamsBody{
			IdOfMaster: int32(n.idOfMaster),
			MyPRecord:  n.myPRecord,
			Verbose:    int32(n.verbose),
		}
		return &ControlMessage{Type: int32(Okay), ParamsBody: &pBody}, nil
	case int32(UpdateParams):
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(AddPeer):
		pRecords := cMsg.ParamsBody.PeerRecords
		n.mergePeerRecords(pRecords)
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(DelPeer):
		pRecords := cMsg.ParamsBody.PeerRecords
		for _, pRec := range pRecords {
			n.deletePeerRecord(int(pRec.Id))
		}
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(Message):
		nMessage := NodeMessage{
			FromPRecord:  n.myPRecord,
			ToID:         cMsg.Spare,
			Type:         int32(Standard),
			StandardBody: &StandardBody{Message: cMsg.Comment}}
		n.DispatchMessage(n.getPeerRecord(int(cMsg.Spare), false), &nMessage)
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(RawMessage):
		n.DispatchMessage(n.getPeerRecord(int(cMsg.NodeMessage.ToID), false), cMsg.NodeMessage)
		return &ControlMessage{Type: int32(Okay)}, nil
	case int32(JoinNetwork):
		if !n.isOnline {
			n.onlineNode()
			return &ControlMessage{Type: int32(Okay)}, nil
		} else {
			return &ControlMessage{Type: int32(Error), Comment: "Node is already online!"}, nil
		}
	case int32(LeaveNetwork):
		if n.isOnline {
			n.offlineNode()
			return &ControlMessage{Type: int32(Okay)}, nil
		} else {
			return &ControlMessage{Type: int32(Error), Comment: "Node is already offline!"}, nil
		}
	case int32(StartElection):
		n.spoofElection()
		return &ControlMessage{Type: int32(Okay)}, nil
	}

	return &ControlMessage{Type: int32(Error), Comment: "Unsupported function."}, nil
}

//Shutdown: Unimplemented as I found that a graceful shutdown is unimportant for this.
func (n *Node) Shutdown(ctx context.Context, cMsg *ControlMessage) (*ControlMessage, error) {
	if cMsg.Type != int32(StopListening) {
		return &ControlMessage{Type: int32(Error), Comment: "Wrong channel."}, nil
	}
	return &ControlMessage{Type: int32(Okay)}, nil
}

//<> "External" Peer-to-Peer API <>

//KeepAlive: Used to check if a node is alive or not.
func (n *Node) KeepAlive(ctx context.Context, inMsg *NodeMessage) (*NodeMessage, error) {
	if !n.isOnline {
		return &NodeMessage{Type: int32(Empty)}, nil
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		return &NodeMessage{FromPRecord: n.myPRecord, Type: int32(NotInNetwork)}, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case int32(KeepAlive):
		nMsg := NodeMessage{FromPRecord: n.myPRecord, Type: int32(KeepAlive), Comment: inMsg.Comment}
		return &nMsg, nil
	default:
		//Received a non-keepalive message on the keepalive channel
		return &NodeMessage{Type: int32(Error), Comment: "Unsupported function."}, nil
	}
}

//SendMessage: Channel for NodeMessages
func (n *Node) SendMessage(ctx context.Context, inMsg *NodeMessage) (*NodeMessage, error) {
	if !n.isOnline {
		return &NodeMessage{Type: int32(Empty)}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Message:", inMsg)
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		nMsg := NodeMessage{FromPRecord: n.myPRecord, Type: int32(NotInNetwork)}
		nMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
		nMsg.Spare = int32(n.idOfMaster)
		return &nMsg, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case int32(Standard):
		fmt.Println(inMsg.FromPRecord.Id, ": ", inMsg.StandardBody.Message)
		return &NodeMessage{FromPRecord: n.myPRecord, Type: int32(Ack)}, nil
	default:
		//Received a non-standard message on the standard channel
		//Consider returning an error or doing something else
		return &NodeMessage{FromPRecord: n.myPRecord, Type: int32(Warning), Comment: "Message sent to wrong channel."}, nil
	}
}

//SendCoordinationMessage: Channgel for Coordination Messages
func (n *Node) SendCoordinationMessage(ctx context.Context, coMsg *CoordinationMessage) (*CoordinationMessage, error) {
	if !n.isOnline {
		return &CoordinationMessage{Type: int32(Empty)}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Coordination Message:", coMsg)
	}
	nCoMsg := CoordinationMessage{FromPRecord: n.myPRecord, Type: int32(Error)}
	//Node is NOT the master and it received an out of network message
	if !n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case int32(ReqToJoin): //Send redirection
			coordinatorRecord := []*PeerRecord{n.getPeerRecord(n.idOfMaster, false)}
			nCoMsg.Type = int32(RedirectToCoordinator)
			nCoMsg.PeerRecords = coordinatorRecord
		default:
			nCoMsg.Type = int32(NotInNetwork)
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is NOT the master and it received an in-network message
	} else if !n.IsMaster() {
		switch coMsg.Type {
		case int32(ApptNewCoordinator): //Accept if message was from master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.idOfMaster = int(coMsg.Spare)
				nCoMsg.Type = int32(Ack)
				if n.IsMaster() {
					n.deletePeerRecord(int(coMsg.FromPRecord.Id))
					n.BroadcastPeerInformation()
				}
			} else {
				nCoMsg.Type = int32(NotMaster)
				nCoMsg.Spare = int32(n.idOfMaster)
			}
		default:
			nCoMsg.Comment = "Unsupported function."
			nCoMsg.Type = int32(Error)
		}
		//Node is master and it received an out of network message
	} else if n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case int32(ReqToJoin):
			pRec := n.getPeerRecord(int(coMsg.FromPRecord.Id), true)
			if pRec == nil { //Allow node to join, send peer records
				n.updatePeerRecords(coMsg.FromPRecord)
				nCoMsg.Type = int32(PeerInformation)
				nCoMsg.PeerRecords = append(n.peerRecords, n.myPRecord)
				//Broadcast info to all nodes
				n.BroadcastPeerInformation()
			} else { //ID collision, REJECT node
				nCoMsg.Type = int32(RejectJoin)
				nCoMsg.Comment = "ID Collision"
			}
		case int32(ReqToMerge):
			nCoMsg.Comment = "Unimplemented funcion."
		default:
			nCoMsg.Type = int32(NotInNetwork)
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is master and it received an in-network message
	} else if n.IsMaster() {
		switch coMsg.Type {
		case int32(ReqToLeave):
			nCoMsg.Type = int32(Ack)
			n.deletePeerRecord(int(coMsg.FromPRecord.Id))
			n.BroadcastPeerInformation()
		case int32(BadNodeReport):
			if n.badNodeHandler(n.getPeerRecord(int(coMsg.Spare), false)) {
				//Master node could reach allegedly downed node. Send peer information to
				//node that made the report in case it was a case of bad network details.
				go n.DispatchCoordinationMessage(coMsg.FromPRecord, &CoordinationMessage{Type: int32(PeerInformation), PeerRecords: append(n.peerRecords, n.myPRecord)})
			}
			nCoMsg.Type = int32(Ack)
		default:
			coMsg.Comment = "Unsupported operation."
		}
	}
	//Messages that apply for in-network nodes regardless of master status
	if n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case int32(PeerInformation): //Accept only if it came from the master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.overridePeerRecords(coMsg.PeerRecords)
			} else if n.IsMaster() { //Some other node think its the master, fix by re-triggering election
				n.spoofElection()
			} else { //The sendee node should not be sending a master-only message
				nCoMsg.Comment = "Sendee node is not coordinator."
				nCoMsg.Type = int32(NotMaster)
			}
		case int32(ElectionResult), int32(ElectSelf), int32(RejectElect):
			n.electionHandler(coMsg)
		}
	}
	return &nCoMsg, nil
}

//SendClientMessage: Channel for ClientMessages
func (n *Node) SentClientMessage(ctx context.Context, CliMsg *ClientMessage) (*ClientMessage, error) {

	var ans int32

	// Replies with master address
	switch CliMsg.Type {
	case int32(FindMaster):
		// Find master
		ans = 5
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)
	case int32(FileRead):
		// Client request read
		ans = 6
		fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)
	case int32(FileWrite):
		ans = 7
		fmt.Printf("> Client %d requesting to write\n", CliMsg.ClientID)
	}
	return &ClientMessage{ClientID: CliMsg.ClientID, Type: int32(Ack), Message: int32(ans)}, nil
}
