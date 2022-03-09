package events

import (
	nc "assignment1/main/NodeComm"
	"fmt"
)

// Keep track of clients who have subscribed to an event

// file_content_modification_subscribers map key: string - file name
// file_content_modification_subscribers map val: []PeerRecord - list of subscribed client addresses
var file_content_modification_subscribers map[string][]*nc.PeerRecord = make(map[string][]*nc.PeerRecord)

// master_fail_over_subscribers val: []PeerRecord - list of subscribed client addresses
var master_fail_over_subscribers []*nc.PeerRecord

// lock_acquisition_subscribers map key: string - lock name
// lock_acquisition_subscribers map val: []PeerRecord - list of subscribed client addresses
var lock_acquisition_subscribers map[string][]*nc.PeerRecord = make(map[string][]*nc.PeerRecord)

// conflicting_lock_request_subscribers map key: string - lock name
// conflicting_lock_request_subscribers map val: []PeerRecord - list of subscribed client addresses
var conflicting_lock_request_subscribers map[string][]*nc.PeerRecord = make(map[string][]*nc.PeerRecord)

// This function subscribes a client to the event in which a
// specific file is modified.
func Subscribe_file_content_modification(client_record *nc.PeerRecord, watching_file string) {
	_, exists := file_content_modification_subscribers[watching_file]
	if exists {
		file_content_modification_subscribers[watching_file] = append(file_content_modification_subscribers[watching_file], client_record)
	} else {
		file_content_modification_subscribers[watching_file] = []*nc.PeerRecord{client_record}
	}
}

// This function subscribes a client to the event in which
// the master fails.
func Subscribe_master_fail_over(client_record *nc.PeerRecord) {
	master_fail_over_subscribers = append(master_fail_over_subscribers, client_record)
}

// This function subscribes a client to the event in which a lock
// that a client is listening for gets aquired by a different client.
func Subscribe_lock_aquisition(client_record *nc.PeerRecord, watching_lock string) {
	_, exists := lock_acquisition_subscribers[watching_lock]
	if exists {
		lock_acquisition_subscribers[watching_lock] = append(lock_acquisition_subscribers[watching_lock], client_record)
	} else {
		lock_acquisition_subscribers[watching_lock] = []*nc.PeerRecord{client_record}
	}
}

// Thic function subscribes a client to the event in which a lock
// that a client currently has is requested by a different client.
func Subscribe_conflicting_lock_request(client_record *nc.PeerRecord, watching_lock string) {
	_, exists := conflicting_lock_request_subscribers[watching_lock]
	if exists {
		conflicting_lock_request_subscribers[watching_lock] = append(conflicting_lock_request_subscribers[watching_lock], client_record)
	} else {
		conflicting_lock_request_subscribers[watching_lock] = []*nc.PeerRecord{client_record}
	}
}

func TestingSubscriptions() {
	client_peer_records := []*nc.PeerRecord{
		{Address: "address1", Port: "port1"},
		{Address: "address2", Port: "port2"},
		{Address: "address3", Port: "port3"},
		{Address: "address4", Port: "port4"},
		{Address: "address5", Port: "port5"},
	}

	things := []string{"thing1", "thing2"}

	for i, client_record := range client_peer_records {
		Subscribe_file_content_modification(client_record, things[i%2])
		Subscribe_master_fail_over(client_record)
		Subscribe_lock_aquisition(client_record, things[i%2])
		Subscribe_conflicting_lock_request(client_record, things[i%2])
	}

	fmt.Println("PRINTING SUBSCRIPTION LISTS")

	fmt.Println("\nfile_content_modification_subscribers:")
	print_subscription_map(file_content_modification_subscribers)
	fmt.Println("\nlock_acquisition_subscribers:")
	print_subscription_map(lock_acquisition_subscribers)
	fmt.Println("\nconflicting_lock_request_subscribers:")
	print_subscription_map(conflicting_lock_request_subscribers)
	fmt.Println("\nmaster_fail_over_subscribers:")
	print_record_array(master_fail_over_subscribers)
	fmt.Print("\n")
}

func print_subscription_map(subscription_map map[string][]*nc.PeerRecord) {
	for key := range subscription_map {
		fmt.Println("key:", key)
		print_record_array(subscription_map[key])
		fmt.Print("\n")
	}
}

func print_record_array(records []*nc.PeerRecord) {
	for _, record := range records {
		fmt.Print(record.Address, ":", record.Port, ", ")
	}
}
