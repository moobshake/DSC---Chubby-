package nodecomm

import (
	"fmt"

	pc "assignment1/main/protocchubby"
)

// EventClientTracker keeps track of clients who have subscribed to an event
type EventClientTracker struct {
	// file_content_modification_subscribers map key: string - file name
	// file_content_modification_subscribers map val: []PeerRecord - list of subscribed client addresses
	file_content_modification_subscribers map[string][]*pc.PeerRecord

	// master_fail_over_subscribers val: []PeerRecord - list of subscribed client addresses
	master_fail_over_subscribers []*pc.PeerRecord

	// lock_acquisition_subscribers map key: string - lock name
	// lock_acquisition_subscribers map val: []PeerRecord - list of subscribed client addresses
	lock_acquisition_subscribers map[string][]*pc.PeerRecord

	// conflicting_lock_request_subscribers map key: string - lock name
	// conflicting_lock_request_subscribers map val: []PeerRecord - list of subscribed client addresses
	conflicting_lock_request_subscribers map[string][]*pc.PeerRecord
}

// This function subscribes a client to the event in which a
// specific file is modified.
func (n *Node) subscribeFileContentModification(clientRecord *pc.PeerRecord, watchingFile string) {
	if n.eventClientTracker.file_content_modification_subscribers == nil {
		n.eventClientTracker.file_content_modification_subscribers = make(map[string][]*pc.PeerRecord)
	}
	_, exists := n.eventClientTracker.file_content_modification_subscribers[watchingFile]
	if exists {
		n.eventClientTracker.file_content_modification_subscribers[watchingFile] = append(n.eventClientTracker.file_content_modification_subscribers[watchingFile], clientRecord)
	} else {
		n.eventClientTracker.file_content_modification_subscribers[watchingFile] = []*pc.PeerRecord{clientRecord}
	}
}

// This function subscribes a client to the event in which
// the master fails.
func (n *Node) subscribeMasterFailOver(clientRecord *pc.PeerRecord) {
	n.eventClientTracker.master_fail_over_subscribers = append(n.eventClientTracker.master_fail_over_subscribers, clientRecord)
}

// This function subscribes a client to the event in which a lock
// that a client is listening for gets aquired by a different client.
func (n *Node) subscribeLockAcquisition(clientRecord *pc.PeerRecord, watchingLock string) {
	if n.eventClientTracker.lock_acquisition_subscribers == nil {
		n.eventClientTracker.lock_acquisition_subscribers = make(map[string][]*pc.PeerRecord)
	}
	_, exists := n.eventClientTracker.lock_acquisition_subscribers[watchingLock]
	if exists {
		n.eventClientTracker.lock_acquisition_subscribers[watchingLock] = append(n.eventClientTracker.lock_acquisition_subscribers[watchingLock], clientRecord)
	} else {
		n.eventClientTracker.lock_acquisition_subscribers[watchingLock] = []*pc.PeerRecord{clientRecord}
	}
}

// Thic function subscribes a client to the event in which a lock
// that a client currently has is requested by a different client.
func (n *Node) subscribeConflictingLockRequest(clientRecord *pc.PeerRecord, watchingLock string) {
	if n.eventClientTracker.conflicting_lock_request_subscribers == nil {
		n.eventClientTracker.conflicting_lock_request_subscribers = make(map[string][]*pc.PeerRecord)
	}
	_, exists := n.eventClientTracker.conflicting_lock_request_subscribers[watchingLock]
	if exists {
		n.eventClientTracker.conflicting_lock_request_subscribers[watchingLock] = append(n.eventClientTracker.conflicting_lock_request_subscribers[watchingLock], clientRecord)
	} else {
		n.eventClientTracker.conflicting_lock_request_subscribers[watchingLock] = []*pc.PeerRecord{clientRecord}
	}
}

func (n *Node) PrintSubscriptionMap(subscriptionMap map[string][]*pc.PeerRecord) {
	for key := range subscriptionMap {
		fmt.Println("key:", key)
		n.PrintRecordArray(subscriptionMap[key])
		fmt.Print("\n")
	}
}

func (n *Node) PrintRecordArray(records []*pc.PeerRecord) {
	for _, record := range records {
		fmt.Print(record.Address, ":", record.Port, ", ")
	}
}

//ClientSubscritipnHandler is the handler for subscriptions - kinda like the accountant of subscriptions
func (n *Node) ClientSubscriptionsHandler(CliMsg *pc.ClientMessage) {
	switch CliMsg.Type {
	case pc.ClientMessage_SubscribeMasterFailover:
		n.subscribeMasterFailOver(CliMsg.ClientAddress)
	case pc.ClientMessage_SubscribeFileModification:
		n.subscribeFileContentModification(CliMsg.ClientAddress, CliMsg.StringMessages)
	case pc.ClientMessage_SubscribeLockAquisition:
		n.subscribeLockAcquisition(CliMsg.ClientAddress, CliMsg.StringMessages)
	case pc.ClientMessage_SubscribeLockConflict:
		n.subscribeConflictingLockRequest(CliMsg.ClientAddress, CliMsg.StringMessages)
	}
}
