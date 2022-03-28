// These events can be subscribed by the clients.
// When the events occur, the subscribed clients will be notified asynchronously

package nodecomm

import (
	"fmt"

	pc "assignment1/main/protocchubby"
)

// This function publishes the event in which a specific file is modified.
func (n *Node) PublishFileContentModification(modified_file string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.file_content_modification_subscribers[modified_file]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publishMessage(clients, pc.EventMessage_FileContentModified, modified_file)
}

// PublishMasterFailover publishes the event in which the master fails.
func (n *Node) PublishMasterFailover() {
	if len(n.eventClientTracker.master_fail_over_subscribers) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publishMessage(n.eventClientTracker.master_fail_over_subscribers, pc.EventMessage_MasterFailOver, "")
}

// PublishLockAcquisition publishes the event in which a lock
// that a client is listening for gets aquired by a different client.
func (n *Node) PublishLockAcquisition(desiredLock string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.lock_acquisition_subscribers[desiredLock]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publishMessage(clients, pc.EventMessage_LockAquisition, desiredLock)
}

// PublishConflictingLockRequest publishes the event in which a lock
// that a client currently has is requested by a different client.
func (n *Node) PublishConflictingLockRequest(desiredLock string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.conflicting_lock_request_subscribers[desiredLock]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publishMessage(clients, pc.EventMessage_ConflictingLock, desiredLock)
}

// func TestPublications() {
// 	things := []string{"thing1", "thing2"}
// 	PublishMasterFailover()

// 	for _, thing := range things {
// 		PublishFileContentModification(thing)
// 		PublishLockAcquisition(thing)
// 		PublishConflictingLockRequest(thing)
// 	}

// 	fmt.Scanln() // press something to stop the test
// }
