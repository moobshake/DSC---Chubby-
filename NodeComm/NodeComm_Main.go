package nodecomm

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	pc "assignment1/main/protocchubby"
)

const (
	LOCAL_ROOT_PATH = "."
	// data{n} where n is the node ID eg data1
	LOCAL_DATA_DIR_PREFIX = "data/data"
	// lock{n} where n is the node ID eg lock1
	LOCAL_LOCK_DIR_PREFIX = "lock/lock"
)

//Node is a logical structure for the bully node.
type Node struct {
	pc.NodeCommControlServiceServer
	pc.NodeCommPeerServiceServer
	pc.NodeCommListeningServiceServer
	idOfMaster           int
	peerRecords          []*pc.PeerRecord
	peerRecordsLock      sync.Mutex
	myPRecord            *pc.PeerRecord
	isOnline             bool
	electionStatus       *pc.ElectionStatus
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
		myPRecord:            &pc.PeerRecord{Id: int32(id), Address: ipAddr, Port: port},
		electionStatus:       &pc.ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)},
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
	pBody := pc.ParamsBody{MyPRecord: n.myPRecord, IdOfMaster: int32(n.idOfMaster), ElectionStatus: n.electionStatus, Verbose: int32(n.verbose), LockGenerationNumber: int32(n.lockGenerationNumber), NodeDataPath: n.nodeDataPath, NodeLockPath: n.nodeLockPath}
	n.DispatchControlMessage(&pc.ControlMessage{Type: pc.ControlMessage_InitParams, ParamsBody: &pBody})
	time.Sleep(time.Second * 1)
	go n.LockChecker()        // start lock checker service
	go n.KeepAliveService(30) // start keep alie service
	go n.MirrorService(60)    // start mirroring service (attempts to mirror every 60 seconds)
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
	pc.RegisterNodeCommControlServiceServer(gServer, &s)
	pc.RegisterNodeCommPeerServiceServer(gServer, &s)
	pc.RegisterNodeCommListeningServiceServer(gServer, &s)

	if err := gServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}
