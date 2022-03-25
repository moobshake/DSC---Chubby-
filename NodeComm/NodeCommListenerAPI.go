package nodecomm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

const (
	LOCAL_ROOT_PATH = "."
	// data{n} where n is the node ID eg data1
	LOCAL_DATA_DIR_PREFIX = "data"
	// lock{n} where n is the node ID eg lock1
	LOCAL_LOCK_DIR_PREFIX = "lock"
)

//Node is a logical structure for the bully node.
type Node struct {
	NodeCommServiceServer
	idOfMaster           int
	peerRecords          []*PeerRecord
	peerRecordsLock      sync.Mutex
	myPRecord            *PeerRecord
	isOnline             bool
	electionStatus       *ElectionStatus
	electionStatusLock   sync.Mutex
	verbose              int
	lockGenerationNumber int
	nodeDataPath         string
	nodeLockPath         string

	eventClientTracker EventClientTracker
}

//CreateNode initialises a Node
func CreateNode(id, idOfMaster int, ipAddr, port string, verbose int) *Node {
	n := Node{
		idOfMaster:           idOfMaster,
		myPRecord:            &PeerRecord{Id: int32(id), Address: ipAddr, Port: port},
		electionStatus:       &ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)},
		verbose:              verbose,
		lockGenerationNumber: 0,
		nodeDataPath:         filepath.Join(LOCAL_ROOT_PATH, LOCAL_DATA_DIR_PREFIX+strconv.Itoa(id)),
		nodeLockPath:         filepath.Join(LOCAL_ROOT_PATH, LOCAL_LOCK_DIR_PREFIX+strconv.Itoa(id)),
	}

	InitDirectory(n.nodeDataPath, true)
	InitDirectory(n.nodeLockPath, false)
	InitLockFiles(n.nodeLockPath, n.nodeDataPath)
	return &n
}

//StartNode starts the listener, initialises the params of the listener, and starts the UI.
func (n *Node) StartNode() {
	go n.startListener()
	pBody := ParamsBody{MyPRecord: n.myPRecord, IdOfMaster: int32(n.idOfMaster), ElectionStatus: n.electionStatus, Verbose: int32(n.verbose), LockGenerationNumber: int32(n.lockGenerationNumber), NodeDataPath: n.nodeDataPath, NodeLockPath: n.nodeLockPath}
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
		n.lockGenerationNumber = int(cMsg.ParamsBody.LockGenerationNumber)
		n.nodeDataPath = cMsg.ParamsBody.NodeDataPath
		n.nodeLockPath = cMsg.ParamsBody.NodeLockPath
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

	// Events
	case ControlMessage_StartElection:
		n.spoofElection()
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_SubscribeMasterFailover:
		n.subscribe_master_fail_over(cMsg.ParamsBody.MyPRecord)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_SubscribeFileModification:
		n.subscribe_file_content_modification(cMsg.ParamsBody.MyPRecord, cMsg.Comment)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_SubscribeLockAquisition:
		n.subscribe_lock_aquisition(cMsg.ParamsBody.MyPRecord, cMsg.Comment)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_SubscribeLockConflict:
		n.subscribe_conflicting_lock_request(cMsg.ParamsBody.MyPRecord, cMsg.Comment)
		return &ControlMessage{Type: ControlMessage_Okay}, nil
	case ControlMessage_PublishMasterFailover:
		n.Publish_master_fail_over()
	case ControlMessage_PublishFileModification:
		n.Publish_file_content_modification(cMsg.Comment)
	case ControlMessage_PublishLockAquisition:
		n.Publish_lock_aquisition(cMsg.Comment)
	case ControlMessage_PublishLockConflict:
		n.Publish_conflicting_lock_request(cMsg.Comment)
	}
	//TODO: Add here

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

//SendCoordinationMessage: Channel for Coordination Messages
func (n *Node) SendCoordinationMessage(ctx context.Context, coMsg *CoordinationMessage) (*CoordinationMessage, error) {
	if !n.isOnline {
		if coMsg.Type == CoordinationMessage_WakeUpAndJoinNetwork {
			fmt.Println("This node has received a WAKEUP message from a coordinator.")
			//LOOK HERE
			n.onlineNode()
		}
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
				nCoMsg.Comment = "Sender node is not coordinator."
				nCoMsg.Type = CoordinationMessage_NotMaster
			}
		case CoordinationMessage_ElectionResult, CoordinationMessage_ElectSelf, CoordinationMessage_RejectElect:
			n.electionHandler(coMsg)
		}
	}
	return &nCoMsg, nil
}

// SendClientMessage: Channel for ClientMessages
// Note: Read Requests are not processed here
func (n *Node) SendClientMessage(ctx context.Context, CliMsg *ClientMessage) (*ClientMessage, error) {

	var ans int32
	var nodeReply string

	//If this server is not the master
	if !(n.IsMaster()) {
		switch CliMsg.Type {
		case ClientMessage_FindMaster:
			n.DispatchClientMessage(CliMsg.ClientAddress, &ClientMessage{Type: ClientMessage_RedirectToCoordinator, Spare: int32(n.idOfMaster), ClientAddress: n.getPeerRecord(n.idOfMaster, true)})
		default:
			n.DispatchClientMessage(CliMsg.ClientAddress, &ClientMessage{Type: ClientMessage_RedirectToCoordinator, Spare: int32(n.idOfMaster), ClientAddress: n.getPeerRecord(n.idOfMaster, true)})
		}
	}

	// Replies with master address
	switch CliMsg.Type {
	case ClientMessage_FindMaster:
		// Find master
		ans = 5
		fmt.Printf("Client %d looking for master\n", CliMsg.ClientID)

	case ClientMessage_FileWrite:
		ans = 7
		fmt.Printf("> Client %d requesting to %s\n", CliMsg.ClientID, CliMsg.Type)

	case ClientMessage_ReadLock:
		isAvail, seq := n.AcquireReadLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}

	case ClientMessage_WriteLock:
		isAvail, seq := n.AcquireWriteLock(CliMsg.StringMessages, int(CliMsg.ClientID), 5)
		if isAvail {
			nodeReply = seq
		} else {
			nodeReply = "NotAvail"
		}

	case ClientMessage_SubscribeFileModification:
		ans = 8
		fmt.Printf("> Client %d requesting to subscibe: %s\n", CliMsg.ClientID, CliMsg.Type.String())
		n.dispatchSubscriptionMessage(CliMsg, ControlMessage_SubscribeFileModification, CliMsg.StringMessages)
	case ClientMessage_SubscribeLockAquisition:
		ans = 9
		fmt.Printf("> Client %d requesting to subscibe: %s\n", CliMsg.ClientID, CliMsg.Type.String())
		n.dispatchSubscriptionMessage(CliMsg, ControlMessage_SubscribeLockAquisition, CliMsg.StringMessages)
	case ClientMessage_SubscribeLockConflict:
		ans = 10
		fmt.Printf("> Client %d requesting to subscibe: %s\n", CliMsg.ClientID, CliMsg.Type.String())
		n.dispatchSubscriptionMessage(CliMsg, ControlMessage_SubscribeLockConflict, CliMsg.StringMessages)
	case ClientMessage_SubscribeMasterFailover:
		ans = 11
		fmt.Printf("> Client %d requesting to subscibe: %s\n", CliMsg.ClientID, CliMsg.Type.String())
		n.dispatchSubscriptionMessage(CliMsg, ControlMessage_SubscribeMasterFailover, "")
	case ClientMessage_ListFile:
		ans = 12
		fmt.Printf("> Client %d requesting to list files\n", CliMsg.ClientID)
		f := list_files(n.nodeDataPath)
		return &ClientMessage{ClientID: CliMsg.ClientID, Type: ClientMessage_Ack, Message: int32(ans), StringMessages: f}, nil
	default:
		ans = -1
		fmt.Printf("> Client %d requesting for something that is not available %s\n", CliMsg.ClientID, CliMsg.Type.String())
	}

	return &ClientMessage{ClientID: CliMsg.ClientID, Type: ClientMessage_Error, Message: int32(ans), StringMessages: nodeReply}, nil
}

// Used to dispatch control messages regarding client event subscriptions.
// This is used when the server listener receives a client subscription request.
func (n *Node) dispatchSubscriptionMessage(CliMsg *ClientMessage, msgType ControlMessage_MessageType, fileLockName string) {
	cMsg := ControlMessage{Type: msgType, Comment: fileLockName, ParamsBody: &ParamsBody{}}
	cMsg.ParamsBody.MyPRecord = &PeerRecord{Address: CliMsg.ClientAddress.Address, Port: CliMsg.ClientAddress.Port}
	n.DispatchControlMessage(&cMsg)
}

func (n *Node) DispatchClientMessage(destPRec *PeerRecord, CliMsg *ClientMessage) *ClientMessage {
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)
	response, err := c.SendClientMessage(context.Background(), CliMsg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}
	return response
}

// Stream the required file from local data to the client in batches
func (n *Node) RequestReadFile(CliMsg *ClientMessage, stream NodeCommService_RequestReadFileServer) error {
	fmt.Printf("> Client %d requesting to read\n", CliMsg.ClientID)

	if n.validateReadRequest(CliMsg) {
		// Get the file from the local dir in batches
		localFilePath := filepath.Join(n.nodeDataPath, CliMsg.StringMessages)
		file, err := os.Open(localFilePath)
		if err != nil {
			fmt.Println("CLIENT FILE READ REQUEST ERROR:", err)
		}
		defer file.Close()

		buffer := make([]byte, READ_MAX_BYTE_SIZE)

		for {
			numBytes, err := file.Read(buffer)

			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
			fileContent := FileBodyMessage{
				Type:        FileBodyMessage_ReadMode,
				FileName:    CliMsg.StringMessages,
				FileContent: buffer[:numBytes],
			}

			if err := stream.Send(&fileContent); err != nil {
				return err
			}
		}

		return nil
	} else {
		// Return an error
		fileContent := FileBodyMessage{
			Type: FileBodyMessage_Error,
		}
		if err := stream.Send(&fileContent); err != nil {
			return err
		}
		return nil
	}

}
