// These events can be subscribed by the clients.
// When the events occur, the subscribed clients will be notified asynchronously

package nodecomm

import (
	"context"
	"fmt"
)

// This function publishes the event in which a specific file
// is modified.
func (n *Node) Publish_file_content_modification(modified_file string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.file_content_modification_subscribers[modified_file]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publish_message(clients, EventMessage_FileContentModified, modified_file)
}

// This function publishes the event in which the master fails.
func (n *Node) Publish_master_fail_over() {
	if len(n.eventClientTracker.master_fail_over_subscribers) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publish_message(n.eventClientTracker.master_fail_over_subscribers, EventMessage_MasterFailOver, "")
}

// This function publishes the event in which a lock
// that a client is listening for gets aquired by a different client.
func (n *Node) Publish_lock_aquisition(desired_lock string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.lock_acquisition_subscribers[desired_lock]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publish_message(clients, EventMessage_LockAquisition, desired_lock)
}

// Thic function publishes the event in which a lock
// that a client currently has is requested by a different client.
func (n *Node) Publish_conflicting_lock_request(desired_lock string) {
	// get all subscribed clients
	clients, exist := n.eventClientTracker.conflicting_lock_request_subscribers[desired_lock]
	if !exist || len(clients) == 0 {
		fmt.Println("No Clients to Publish To")
		return
	}
	go n.publish_message(clients, EventMessage_ConflictingLock, desired_lock)
}

// Publish all messages to the subscribed clients
func (n *Node) publish_message(subscribed_clients []*PeerRecord, eventType EventMessage_MessageType, fileOrLockName string) {
	for _, clientRecord := range subscribed_clients {
		fmt.Println("TEST sending client", clientRecord.Address, clientRecord.Port, ": Event Type", eventType.String(), fileOrLockName)
		n.publishEventMessage(clientRecord, eventType, fileOrLockName)
	}

}

// server publishes event messages to the clients upon event happenings
func (n *Node) publishEventMessage(clientRecord *PeerRecord, eventType EventMessage_MessageType, fileOrLockName string) {
	conn, err := connectTo(clientRecord.Address, clientRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := NewNodeCommServiceClient(conn)

	var msg EventMessage

	switch eventType {
	case EventMessage_FileContentModified:
		msg = EventMessage{Type: EventMessage_FileContentModified, FileName: fileOrLockName}
	case EventMessage_LockAquisition:
		msg = EventMessage{Type: EventMessage_LockAquisition, LockName: fileOrLockName}
	case EventMessage_ConflictingLock:
		msg = EventMessage{Type: EventMessage_ConflictingLock, LockName: fileOrLockName}
	case EventMessage_MasterFailOver:
		msg = EventMessage{Type: EventMessage_MasterFailOver}
	}

	response, err := c.SendEventMessage(context.Background(), &msg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}

	if response.Type != EventMessage_Ack {
		fmt.Println("Expecting event response", EventMessage_Ack, "from event -", eventType.String(), "but got", response.Type)
	}
}

// func TestPublications() {
// 	things := []string{"thing1", "thing2"}
// 	Publish_master_fail_over()

// 	for _, thing := range things {
// 		Publish_file_content_modification(thing)
// 		Publish_lock_aquisition(thing)
// 		Publish_conflicting_lock_request(thing)
// 	}

// 	fmt.Scanln() // press something to stop the test
// }
