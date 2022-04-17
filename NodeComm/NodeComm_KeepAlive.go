package nodecomm

import (
	"time"
)

//For replica <-> replica
// Warning: This method assumes that the peer record for the master is correct.
//--> If a message was received recently, consider not dispatching a keepalive
func (n *Node) KeepAliveService(interval int) {
	for {
		time.Sleep(time.Second * time.Duration(interval))
		if n.IsMaster() {
			continue
		}
		if !n.isOnline {
			break
		}
		keepAliveSuccess := n.DispatchKeepAlive(n.getPeerRecord(n.idOfMaster, false))
		if !keepAliveSuccess { // Verify that coordinator is down
			keepAliveSuccess = n.badNodeHandler(n.getPeerRecord(n.idOfMaster, false))
		}
		if !keepAliveSuccess { //Wait for election to finish
			time.Sleep(time.Second * time.Duration(interval))
		}
	}
}

//For master <-> client
func (n *Node) ClientKeepAliveService(interval int) {
	for {
		time.Sleep(time.Second * time.Duration(interval))
		if !n.IsMaster() {
			continue
		}
		if !n.isOnline {
			break
		}
		n.activeClientsLock.Lock()
		*n.activeClients = make([]int, len(*n._activeClients))
		*n.activeClients = *n._activeClients
		*n._activeClients = []int{}
		n.activeClientsLock.Unlock()
	}
}

func (n *Node) NoteClientIsAlive(clientID int) {
	n.activeClientsLock.Lock()
	if !InIntArray(n.activeClients, clientID) {
		*n.activeClients = append(*n.activeClients, int(clientID))
	}
	if !InIntArray(n._activeClients, clientID) {
		*n._activeClients = append(*n._activeClients, int(clientID))
	}
	n.activeClientsLock.Unlock()
}

func (n *Node) isClientActiveLock(id int) bool {
	if !n.IsMaster() { //Darryl level code
		return true
	}
	n.activeClientsLock.Lock()
	result := InIntArray(n.activeClients, id)
	n.activeClientsLock.Unlock()
	return result
}
