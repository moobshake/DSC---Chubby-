package nodecomm

import (
	"fmt"
	"time"
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
	fmt.Println("Node is online.")
}

//tryJoinNetwork has the node contact a configured coordinator to join a network or it has the
//node create a network if it is configured as a coordinator.
func (n *Node) tryJoinNetwork() bool {
	nCoMsg := CoordinationMessage{Type: CoordinationMessage_ReqToJoin}
	response := n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &nCoMsg)
	if response == nil {
		return false
	}
	for response.Type == CoordinationMessage_RedirectToCoordinator {
		fmt.Println("Attempt to join was rejected because specified node is not the master.")
		fmt.Println("Received redirection, attempting to join network of node:", int(response.PeerRecords[0].Id))
		n.mergePeerRecords(response.PeerRecords)
		response = n.DispatchCoordinationMessage(response.PeerRecords[0], &nCoMsg)
		if response == nil {
			return false
		}
	}
	if response.Type == CoordinationMessage_RejectJoin {
		fmt.Println("Attempt to join was rejected: ", response.Comment)
		return false
	} else if response.Type == CoordinationMessage_PeerInformation {
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
		n.BroadcastCoordinationMessage(&CoordinationMessage{Type: CoordinationMessage_ApptNewCoordinator, Spare: int32(nextHighestID)})
		return
	}
	n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &CoordinationMessage{Type: CoordinationMessage_ReqToLeave})
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
func (n *Node) getPeerRecord(ID int, noWarning bool) *PeerRecord {
	if ID == int(n.myPRecord.Id) {
		return n.myPRecord
	}
	var ansRec *PeerRecord
	n.peerRecordsLock.Lock()
	for _, pRec := range n.peerRecords {
		if pRec.Id == int32(ID) {
			ansRec = &PeerRecord{Id: pRec.Id, Address: pRec.Address, Port: pRec.Port}
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
func (n *Node) updatePeerRecords(pUpdateRec *PeerRecord) {
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
func (n *Node) mergePeerRecords(peerRecords []*PeerRecord) {
	for _, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			continue
		}
		n.updatePeerRecords(nPRec)
	}
}

//overrridePeerRecords completely overrides this node's PeerRecords with the given PeerRecords
func (n *Node) overridePeerRecords(peerRecords []*PeerRecord) {
	n.peerRecords = peerRecords
	for i, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			peerRecords = append(peerRecords[:i], peerRecords[i+1:]...)
		}
	}
	n.peerRecords = peerRecords
}

//badNodeHandler checks the node of the given pRec and handles it if it is offline
func (n *Node) badNodeHandler(pRec *PeerRecord) bool {
	//This check is especially important because if a coordinator receives a badnodereport and an election starts
	//simultaneously, it might remove the peerrecord of the badnode and attempt to send an election message
	//to the same node. Racecondition problems. This might cause an nil error if this check isn't here.
	if pRec == nil {
		return false
	}
	fmt.Println("Potentially offline node detected: ", pRec)
	fmt.Println("Sending KeepAlive message...")
	response := n.DispatchKeepAlve(pRec)
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
		n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &CoordinationMessage{Type: CoordinationMessage_BadNodeReport, Spare: pRec.Id})
	}
	return false
}

//electionHandler handles all election messages
func (n *Node) electionHandler(electMsg *CoordinationMessage) {
	n.electionStatusLock.Lock()
	n.electionStatus.Active = 2
	switch electMsg.Type {
	case CoordinationMessage_ElectSelf:
		//If no ongoing election and node lost before it even began
		if n.electionStatus.OngoingElection == 1 && electMsg.FromPRecord.Id > n.myPRecord.Id {
			n.electionStatus.OngoingElection = 2
			n.electionStatus.IsWinning = 1
			n.electionStatusLock.Unlock()
			return
		}
		if n.electionStatus.OngoingElection == 1 { //If no ongoing election and node has a chance of winning
			n.electionStatus.OngoingElection = 2
			n.electionStatus.IsWinning = 2
			go n.electionTimer()
			if electMsg.FromPRecord.Id < n.myPRecord.Id { //Send reject to lower ID nodes
				n.DispatchCoordinationMessage(electMsg.FromPRecord, &CoordinationMessage{Type: CoordinationMessage_RejectElect})
			}
			n.electionStatusLock.Unlock()
			return
		}
		if electMsg.FromPRecord.Id < n.myPRecord.Id { //Send reject to lower ID nodes
			n.DispatchCoordinationMessage(electMsg.FromPRecord, &CoordinationMessage{Type: CoordinationMessage_RejectElect})
		}
	case CoordinationMessage_RejectElect:
		if electMsg.FromPRecord.Id > n.myPRecord.Id {
			n.electionStatus.IsWinning = 1
		}
	case CoordinationMessage_ElectionResult:
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
	for _, pRec := range n.peerRecords {
		if pRec.Id > n.myPRecord.Id {
			n.DispatchCoordinationMessage(pRec, &CoordinationMessage{Type: CoordinationMessage_ElectSelf})
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
	fakeCoMsg := CoordinationMessage{Type: CoordinationMessage_ElectSelf, FromPRecord: n.myPRecord}
	n.electionHandler(&fakeCoMsg)
}

//Utility
func (n *Node) IsMaster() bool {
	return n.myPRecord.Id == int32(n.idOfMaster)
}
