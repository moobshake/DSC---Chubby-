package nodecomm

import (
	"fmt"
	"time"
)

//Methods here MUST ONLY BE CALLED BY THE LISTENER SERVER

//online: API command to have the node join a network
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

func (n *Node) tryJoinNetwork() bool {
	nCoMsg := CoordinationMessage{Type: int32(ReqToJoin)}
	response := n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &nCoMsg)
	if response == nil {
		return false
	}
	for response.Type == int32(RedirectToCoordinator) {
		fmt.Println("Attempt to join was rejected because specified node is not the master.")
		fmt.Println("Received redirection, attempting to join network of node:", int(response.PeerRecords[0].Id))
		n.mergePeerRecords(response.PeerRecords)
		response = n.DispatchCoordinationMessage(response.PeerRecords[0], &nCoMsg)
		if response == nil {
			return false
		}
	}
	if response.Type == int32(RejectJoin) {
		fmt.Println("Attempt to join was rejected: ", response.Comment)
		return false
	} else if response.Type == int32(PeerInformation) {
		numNodes := len(response.PeerRecords)
		fmt.Println("Successfully joined network of node", response.FromPRecord.Id, ". There are ", numNodes-1, " other nodes in the network.")
		n.mergePeerRecords(response.PeerRecords)
		n.idOfMaster = int(response.FromPRecord.Id)
		return true
	}
	return false
}

func (n *Node) offlineNode() {
	if n.idOfMaster == int(n.myPRecord.Id) {
		nextHighestID := 0
		for i := len(n.peerRecords) - 1; i >= 0; i-- {
			if n.peerRecords[i].Id > int32(nextHighestID) {
				nextHighestID = int(n.peerRecords[i].Id)
			}
		}
		n.BroadcastCoordinationMessage(&CoordinationMessage{Type: int32(ApptNewCoordinator), Spare: int32(nextHighestID)})
		return
	}
	n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &CoordinationMessage{Type: int32(ReqToLeave)})
	n.isOnline = false
	fmt.Println("Node is offline.")
}

//Recordkeeping Code
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

//Safe
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

//Safe
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

func (n *Node) mergePeerRecords(peerRecords []*PeerRecord) {
	for _, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			continue
		}
		n.updatePeerRecords(nPRec)
	}
}

func (n *Node) overridePeerRecords(peerRecords []*PeerRecord) {
	n.peerRecords = peerRecords
	for i, nPRec := range peerRecords {
		if nPRec.Id == n.myPRecord.Id {
			peerRecords = append(peerRecords[:i], peerRecords[i+1:]...)
		}
	}
	n.peerRecords = peerRecords
}

func (n *Node) badNodeHandler(pRec *PeerRecord) bool {
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
		n.DispatchCoordinationMessage(n.getPeerRecord(n.idOfMaster, false), &CoordinationMessage{Type: int32(BadNodeReport), Spare: pRec.Id})
	}
	return false
}

//Election code
func (n *Node) electionHandler(electMsg *CoordinationMessage) {
	n.electionStatusLock.Lock()
	n.electionStatus.Active = 2
	switch electMsg.Type {
	case int32(ElectSelf):
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
				n.DispatchCoordinationMessage(electMsg.FromPRecord, &CoordinationMessage{Type: int32(RejectElect)})
			}
			n.electionStatusLock.Unlock()
			return
		}
		if electMsg.FromPRecord.Id < n.myPRecord.Id { //Send reject to lower ID nodes
			n.DispatchCoordinationMessage(electMsg.FromPRecord, &CoordinationMessage{Type: int32(RejectElect)})
		}
	case int32(RejectElect):
		if electMsg.FromPRecord.Id > n.myPRecord.Id {
			n.electionStatus.IsWinning = 1
		}
	case int32(ElectionResult):
		n.idOfMaster = int(electMsg.FromPRecord.Id)
		n.overridePeerRecords(electMsg.PeerRecords)
		n.electionStatus.OngoingElection = 1
		fmt.Println("Notice: New master node is", n.idOfMaster)
	}
	n.electionStatusLock.Unlock()
}

func (n *Node) electionTimer() {
	for _, pRec := range n.peerRecords {
		if pRec.Id > n.myPRecord.Id {
			n.DispatchCoordinationMessage(pRec, &CoordinationMessage{Type: int32(ElectSelf)})
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

func (n *Node) spoofElection() {
	fakeCoMsg := CoordinationMessage{Type: int32(ElectSelf), FromPRecord: n.myPRecord}
	n.electionHandler(&fakeCoMsg)
}

//Utility
func (n *Node) IsMaster() bool {
	return n.myPRecord.Id == int32(n.idOfMaster)
}
