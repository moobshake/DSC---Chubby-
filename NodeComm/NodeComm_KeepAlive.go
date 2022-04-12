package nodecomm

import "time"

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
