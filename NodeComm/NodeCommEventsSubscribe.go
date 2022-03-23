package nodecomm

import (
	"fmt"
)

// Keep track of clients who have subscribed to an event
type EventClientTracker struct {
	// file_content_modification_subscribers map key: string - file name
	// file_content_modification_subscribers map val: []PeerRecord - list of subscribed client addresses
	file_content_modification_subscribers map[string][]*PeerRecord

	// master_fail_over_subscribers val: []PeerRecord - list of subscribed client addresses
	master_fail_over_subscribers []*PeerRecord

	// lock_acquisition_subscribers map key: string - lock name
	// lock_acquisition_subscribers map val: []PeerRecord - list of subscribed client addresses
	lock_acquisition_subscribers map[string][]*PeerRecord

	// conflicting_lock_request_subscribers map key: string - lock name
	// conflicting_lock_request_subscribers map val: []PeerRecord - list of subscribed client addresses
	conflicting_lock_request_subscribers map[string][]*PeerRecord
}

// This function subscribes a client to the event in which a
// specific file is modified.
func (n *Node) subscribe_file_content_modification(client_record *PeerRecord, watching_file string) {
	if n.eventClientTracker.file_content_modification_subscribers == nil {
		n.eventClientTracker.file_content_modification_subscribers = make(map[string][]*PeerRecord)
	}
	_, exists := n.eventClientTracker.file_content_modification_subscribers[watching_file]
	if exists {
		n.eventClientTracker.file_content_modification_subscribers[watching_file] = append(n.eventClientTracker.file_content_modification_subscribers[watching_file], client_record)
	} else {
		n.eventClientTracker.file_content_modification_subscribers[watching_file] = []*PeerRecord{client_record}
	}
}

// This function subscribes a client to the event in which
// the master fails.
func (n *Node) subscribe_master_fail_over(client_record *PeerRecord) {
	n.eventClientTracker.master_fail_over_subscribers = append(n.eventClientTracker.master_fail_over_subscribers, client_record)
}

// This function subscribes a client to the event in which a lock
// that a client is listening for gets aquired by a different client.
func (n *Node) subscribe_lock_aquisition(client_record *PeerRecord, watching_lock string) {
	if n.eventClientTracker.lock_acquisition_subscribers == nil {
		n.eventClientTracker.lock_acquisition_subscribers = make(map[string][]*PeerRecord)
	}
	_, exists := n.eventClientTracker.lock_acquisition_subscribers[watching_lock]
	if exists {
		n.eventClientTracker.lock_acquisition_subscribers[watching_lock] = append(n.eventClientTracker.lock_acquisition_subscribers[watching_lock], client_record)
	} else {
		n.eventClientTracker.lock_acquisition_subscribers[watching_lock] = []*PeerRecord{client_record}
	}
}

// Thic function subscribes a client to the event in which a lock
// that a client currently has is requested by a different client.
func (n *Node) subscribe_conflicting_lock_request(client_record *PeerRecord, watching_lock string) {
	if n.eventClientTracker.conflicting_lock_request_subscribers == nil {
		n.eventClientTracker.conflicting_lock_request_subscribers = make(map[string][]*PeerRecord)
	}
	_, exists := n.eventClientTracker.conflicting_lock_request_subscribers[watching_lock]
	if exists {
		n.eventClientTracker.conflicting_lock_request_subscribers[watching_lock] = append(n.eventClientTracker.conflicting_lock_request_subscribers[watching_lock], client_record)
	} else {
		n.eventClientTracker.conflicting_lock_request_subscribers[watching_lock] = []*PeerRecord{client_record}
	}
}

func (n *Node) Print_subscription_map(subscription_map map[string][]*PeerRecord) {
	for key := range subscription_map {
		fmt.Println("key:", key)
		n.Print_record_array(subscription_map[key])
		fmt.Print("\n")
	}
}

func (n *Node) Print_record_array(records []*PeerRecord) {
	for _, record := range records {
		fmt.Print(record.Address, ":", record.Port, ", ")
	}
}
