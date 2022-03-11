// These events can be subscribed by the clients.
// When the events occur, the subscribed clients will be notified asynchronously

package events

import (
	nc "assignment1/main/NodeComm"
	"fmt"
)

const (
	FILE_CONTENT_MODIFIED_PREFIX = "FILE_CONTENT_MODIFIED"
	MASTER_FAIL_OVER_PREFIX      = "MASTER_FAIL_OVER"
	LOCK_AQUISITION_PREFIX       = "LOCK_AQUISITION"
	LOCK_CONFLICT_PREFIX         = "LOCK_CONFLICT"
)

// This function publishes the event in which a specific file
// is modified.
func Publish_file_content_modification(modified_file string) {
	// get all subscribed clients
	clients, exist := file_content_modification_subscribers[modified_file]
	if !exist || len(clients) == 0 {
		return
	}
	go publish_message(clients, FILE_CONTENT_MODIFIED_PREFIX+", "+modified_file)
}

// This function publishes the event in which the master fails.
func Publish_master_fail_over() {
	if len(master_fail_over_subscribers) == 0 {
		return
	}
	go publish_message(master_fail_over_subscribers, MASTER_FAIL_OVER_PREFIX)
}

// This function publishes the event in which a lock
// that a client is listening for gets aquired by a different client.
func Publish_lock_aquisition(desired_lock string) {
	// get all subscribed clients
	clients, exist := lock_acquisition_subscribers[desired_lock]
	if !exist || len(clients) == 0 {
		return
	}
	go publish_message(clients, LOCK_AQUISITION_PREFIX+", "+desired_lock)
}

// Thic function publishes the event in which a lock
// that a client currently has is requested by a different client.
func Publish_conflicting_lock_request(desired_lock string) {
	// get all subscribed clients
	clients, exist := conflicting_lock_request_subscribers[desired_lock]
	if !exist || len(clients) == 0 {
		return
	}
	go publish_message(clients, LOCK_CONFLICT_PREFIX+", "+desired_lock)
}

// might want to change message into a struct or something, depending on how the message will be structured.
// need to ask YH, I dont want to change the proto file at the same time.
func publish_message(subscribed_clients []*nc.PeerRecord, message string) {
	for _, client_record := range subscribed_clients {
		fmt.Println("TEST sending client", client_record.Address, client_record.Port, ":", message)

		// TODO: UNCOMMENT ONCE THE CLIENT MESSAGES HAVE BEEN FINALISED
		// MIGHT NOT EVEN BE SENDING THESE MESSAGES IN THIS LIB

		// conn, err := nc.connectTo(client_record.Address, client_record.Port)
		// if err != nil {
		// 	fmt.Println("Error connecting:", err)
		// }
		// defer conn.Close()

		// c := nc.NewNodeCommServiceClient(conn)
		// msg := SOME_CLIENT_MESSAGE{}
		// response, err := c.SEND_SOME_CLIENT_MESSGE(context.Background(), &msg)
		// if err != nil {
		// 	fmt.Println("Error dispatching control message:", err)
		// }

		// // ok i know the response is not a string
		// // so this will have to change once I find out what the response is
		// if response != "OK" {
		// 	fmt.Println(response)
		// }
	}

}

func TestPublications() {
	things := []string{"thing1", "thing2"}
	Publish_master_fail_over()

	for _, thing := range things {
		Publish_file_content_modification(thing)
		Publish_lock_aquisition(thing)
		Publish_conflicting_lock_request(thing)
	}

	fmt.Scanln() // press something to stop the test
}
