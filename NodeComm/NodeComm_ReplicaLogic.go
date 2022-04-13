package nodecomm

import (
	"fmt"
	"time"

	pc "assignment1/main/protocchubby"
)

//Methods here MUST ONLY BE CALLED BY THE LISTENER SERVER

//onlineNode: API command to have the node join a network
func (n *Node) onlineNode() {
	if n.idOfMaster != int(n.myPRecord.Id) && n.idOfMaster == -1 {
		fmt.Println("Unable to start or join a network without idOfMaster set!")
		n.isOnline = false
		return
	}
	if n.idOfMaster != int(n.myPRecord.Id) {
		if !n.tryJoinNetwork() {
			fmt.Println("Unable to start as unable to join network of configured master.")
			n.isOnline = false
			return
		}
	}
	n.isOnline = true
	go n.KeepAliveService(30) // start keep alive service
	go n.MirrorService(60)    // start mirroring service (attempts to mirror every 60 seconds)
	go n.LockChecker()        // start lock checker service
	fmt.Println("Node is online.")
}

//tryJoinNetwork has the node contact a configured coordinator to join a network or it has the
//node create a network if it is configured as a coordinator.
func (n *Node) tryJoinNetwork() bool {
	nCoMsg := pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqToJoin}
	response := n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &nCoMsg)
	if response == nil {
		return false
	}
	for response.Type == pc.CoordinationMessage_RedirectToCoordinator {
		fmt.Println("Attempt to join was rejected because specified node is not the master.")
		fmt.Println("Received redirection, attempting to join network of node:", int(response.PeerRecords[0].Id))
		n.mergePeerRecords(response.PeerRecords)
		response = n.DispatchCoordinationMessage(response.PeerRecords[0], &nCoMsg)
		if response == nil {
			return false
		}
	}
	if response.Type == pc.CoordinationMessage_RejectJoin {
		fmt.Println("Attempt to join was rejected: ", response.Comment)
		return false
	} else if response.Type == pc.CoordinationMessage_PeerInformation {
		numNodes := len(response.PeerRecords)
		fmt.Println("Successfully joined network of node", response.FromPRecord.Id, ". There are ", numNodes-1, " other nodes in the network.")
		n.mergePeerRecords(response.PeerRecords)
		n.idOfMaster = int(response.FromPRecord.Id)
		return true
	}
	return false
}

//offlineNode instructs a node to leave its network. If said node is also a master, it appoints
//the next highest ID node to take cover as coordinator.
func (n *Node) offlineNode() {
	if n.idOfMaster == int(n.myPRecord.Id) {
		nextHighestID := 0
		for i := len(n.peerRecords) - 1; i >= 0; i-- {
			if n.peerRecords[i].Id > int32(nextHighestID) {
				nextHighestID = int(n.peerRecords[i].Id)
			}
		}
		n.BroadcastCoordinationMessage(&pc.CoordinationMessage{Type: pc.CoordinationMessage_ApptNewCoordinator, Spare: int32(nextHighestID)})
		return
	}
	n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqToLeave})
	n.isOnline = false
	fmt.Println("Node is offline.")
}

//Recordkeeping Code
//isPeer returns true if the given node ID is known to this node.
func (n *Node) isPeer(senderID int32) bool {
	peerID := int(senderID)
	var ans bool
	n.peerRecordsLock.Lock()
	for _, peerRecord := range n.peerRecords {
		if int(peerRecord.Id) == peerID {
			ans = true
			break
		}
	}
	n.peerRecordsLock.Unlock()
	return ans
}

//getPeerRecord returns the PeerRecord of the given node ID.
func (n *Node) getPeerRecord(ID int, noWarning bool) *pc.PeerRecord {
	if ID == int(n.myPRecord.Id) {
		return n.myPRecord
	}
	var ansRec *pc.PeerRecord
	n.peerRecordsLock.Lock()
	for _, pRec := range n.peerRecords {
		if pRec.Id == int32(ID) {
			ansRec = &pc.PeerRecord{Id: pRec.Id, Address: pRec.Address, Port: pRec.Port}
			break
		}
	}
	n.peerRecordsLock.Unlock()
	if ansRec == nil && !noWarning {
		fmt.Println("Warning: No peer record found for node", ID)
	}
	return ansRec
}

//deletePeerRecord deletes a PeerRecord matching the given ID if such a record exists.
func (n *Node) deletePeerRecord(ID int) {
	n.peerRecordsLock.Lock()
	for i, pRec := range n.peerRecords {
		if pRec.Id == int32(ID) {
			n.peerRecords = append(n.peerRecords[:i], n.peerRecords[i+1:]...)
			break
		}
	}
	n.peerRecordsLock.Unlock()
}

//updatePeerRecords updates the PeerRecords of the node with the given PeerRecord
func (n *Node) updatePeerRecords(pUpdateRec *pc.PeerRecord) {
	n.peerRecordsLock.Lock()
	for i, pRec := range n.peerRecords {
		if pRec.Id == pUpdateRec.Id {
			n.peerRecords[i] = pUpdateRec
			n.peerRecordsLock.Unlock()
			return
		}
	}
	n.peerRecords = append(n.peerRecords, pUpdateRec)
	n.peerRecordsLock.Unlock()
}

//mergePeerRecords updates the PeerRecords of the node with the given PeerRecords
func (n *Node) mergePeerRecords(peerRecords []*pc.PeerRecord) {
	for _, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			continue
		}
		n.updatePeerRecords(nPRec)
	}
}

//overrridePeerRecords completely overrides this node's PeerRecords with the given PeerRecords
func (n *Node) overridePeerRecords(peerRecords []*pc.PeerRecord) {
	n.peerRecords = peerRecords
	for i, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			peerRecords = append(peerRecords[:i], peerRecords[i+1:]...)
		}
	}
	n.peerRecords = peerRecords
}

//badNodeHandler checks the node of the given pRec and handles it if it is offline
func (n *Node) badNodeHandler(pRec *pc.PeerRecord) bool {
	//This check is especially important because if a coordinator receives a badnodereport and an election starts
	//simultaneously, it might remove the peerrecord of the badnode and attempt to send an election message
	//to the same node. Racecondition problems. This might cause an nil error if this check isn't here.
	if pRec == nil {
		return false
	}
	if pRec.Id == int32(0) { // not sure if this works, but basically I need to avoid the below code when the intention is wakeUpNode
		return false
	}
	fmt.Println("Potentially offline node detected: ", pRec)
	fmt.Println("Sending KeepAlive message...")
	response := n.DispatchKeepAlive(pRec)
	if response {
		fmt.Println(pRec, "responded to keepalive signal.")
		return true
	}
	fmt.Println(pRec, "confirmed to be offline.")
	if pRec.Id == int32(n.idOfMaster) { //If downed node was master, start election
		n.deletePeerRecord(int(pRec.Id))
		n.spoofElection()
	} else if n.IsMaster() { //If this node is the master, update the network
		n.deletePeerRecord(int(pRec.Id))
		n.BroadcastPeerInformation()
	} else { //This node is not the master and the downed node is not a master. Inform the master.
		n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &pc.CoordinationMessage{Type: pc.CoordinationMessage_BadNodeReport, Spare: pRec.Id})
	}
	return false
}

//electionHandler handles all election messages
//Bully algorithm - modified to prioritise last outstanding files
func (n *Node) electionHandler(electMsg *pc.CoordinationMessage) {
	n.electionStatusLock.Lock()
	n.electionStatus.Active = 2
	switch electMsg.Type {
	case pc.CoordinationMessage_ElectSelf:

		//If there is no ongoing election
		if n.electionStatus.OngoingElection == 1 {
			n.electionStatus.NumOutstandingFiles = int32(len(n.outstandingFiles))
			var isPossibleToWin bool
			//Only the spoofed message should have the same ID
			if n.myPRecord.Id == electMsg.FromPRecord.Id {
				isPossibleToWin = true
			} else if n.electionStatus.NumOutstandingFiles > electMsg.Spare { //If this node has more outstanding requests, it loses by default
				isPossibleToWin = false
			} else if n.electionStatus.NumOutstandingFiles < electMsg.Spare { //If this node has less outstanding files, it has a chance of winning
				isPossibleToWin = true
			} else if n.electionStatus.NumOutstandingFiles == electMsg.Spare { //If this node has equal outstanding files, break using node ID
				if n.myPRecord.Id > electMsg.FromPRecord.Id {
					isPossibleToWin = true
				} else {
					isPossibleToWin = false
				}
			}
			if isPossibleToWin {
				n.electionStatus.OngoingElection = 2
				n.electionStatus.IsWinning = 2
				go n.electionTimer()
				n.DispatchCoordinationMessage(electMsg.FromPRecord, &pc.CoordinationMessage{Type: pc.CoordinationMessage_RejectElect})
			} else {
				n.electionStatus.OngoingElection = 2
				n.electionStatus.IsWinning = 1
			}
			n.electionStatusLock.Unlock()
			return
		}

		//If there is an ongoing election, send reject if rival node has more outstanding files or if equal, a lower ID
		var sendReject bool
		if n.electionStatus.NumOutstandingFiles > electMsg.Spare { //If this node has more outstanding requests, it loses
			sendReject = false
		} else if n.electionStatus.NumOutstandingFiles < electMsg.Spare { //If this node has less outstanding files, reject
			sendReject = true
		} else if n.electionStatus.NumOutstandingFiles == electMsg.Spare { //If this node has equal outstanding files, break using node ID
			if n.myPRecord.Id > electMsg.FromPRecord.Id {
				sendReject = true
			} else {
				sendReject = false
			}
		}
		if sendReject {
			n.DispatchCoordinationMessage(electMsg.FromPRecord, &pc.CoordinationMessage{Type: pc.CoordinationMessage_RejectElect})
		}
	case pc.CoordinationMessage_RejectElect:
		var acceptReject bool
		if n.electionStatus.NumOutstandingFiles > electMsg.Spare { //If this node has more outstanding requests, it loses
			acceptReject = true
		} else if n.electionStatus.NumOutstandingFiles < electMsg.Spare { //If this node has less outstanding files, reject
			acceptReject = false
		} else if n.electionStatus.NumOutstandingFiles == electMsg.Spare { //If this node has equal outstanding files, break using node ID
			if n.myPRecord.Id > electMsg.FromPRecord.Id {
				acceptReject = false
			} else {
				acceptReject = true
			}
		}
		if acceptReject {
			n.electionStatus.IsWinning = 1
		}
	case pc.CoordinationMessage_ElectionResult:
		n.idOfMaster = int(electMsg.FromPRecord.Id)
		n.overridePeerRecords(electMsg.PeerRecords)
		n.electionStatus.OngoingElection = 1
		fmt.Println("Notice: New master node is", n.idOfMaster)
	}
	n.electionStatusLock.Unlock()
}

//electionTimer is used for a node to check if it has won the election or not.
//When the timer expires, the node will check if it received any RejectElect, if it has not
//it declares itself the winner of the election.
func (n *Node) electionTimer() {
	//Send out electSelf messages
	for _, pRec := range n.peerRecords {
		if pRec.Id > n.myPRecord.Id {
			n.DispatchCoordinationMessage(pRec, &pc.CoordinationMessage{Type: pc.CoordinationMessage_ElectSelf, Spare: int32(len(n.outstandingFiles))})
		}
	}
	//Wait for election to finish
	for n.electionStatus.Active == 2 {
		n.electionStatus.Active = 1
		time.Sleep(time.Second * time.Duration(n.electionStatus.TimeoutDuration))
	}
	if n.electionStatus.IsWinning == 2 { //Node won the election
		fmt.Println("Notice: This node is now the master node.")
		n.idOfMaster = int(n.myPRecord.Id)
		//remove higher ID nodes from my peer record since they are obviously non responsive
		for i := len(n.peerRecords) - 1; i >= 0; i-- {
			if n.peerRecords[i].Id > n.myPRecord.Id {
				n.peerRecords = append(n.peerRecords[:i], n.peerRecords[i+1:]...)
			}
		}
		n.BroadcastElectionResults()
		n.electionStatus.OngoingElection = 1
	} else { //Node lost the election
		//Election results aren't in yet, wait a bit longer
		if n.electionStatus.OngoingElection == 2 {
			return
		}
		time.Sleep(time.Second * time.Duration(n.electionStatus.TimeoutDuration*2))
		//Election results are MIA, did the to-be-coordinator fail? Trigger another election.
		//I might want to assess if the recursion here could pose a problem.
		if n.electionStatus.OngoingElection == 2 {
			fmt.Println("Election results not obtained! Master node unchanged. Re-triggering election.")
			n.electionStatus.OngoingElection = 1
			n.spoofElection()
		}
	}
}

//spoofElection spoofs a fake ElectSelf to this node to prompt it start an election
func (n *Node) spoofElection() {
	fakeCoMsg := pc.CoordinationMessage{Type: pc.CoordinationMessage_ElectSelf, FromPRecord: n.myPRecord}
	n.electionHandler(&fakeCoMsg)
}

//Utility
func (n *Node) IsMaster() bool {
	return n.myPRecord.Id == int32(n.idOfMaster)
}
