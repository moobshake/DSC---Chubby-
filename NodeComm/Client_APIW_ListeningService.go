package nodecomm

import (
	"context"
	"fmt"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

//DispatchClientMessage sends a client message
func (n *Node) DispatchClientMessage(destPRec *pc.PeerRecord, CliMsg *pc.ClientMessage) *pc.ClientMessage {
	if destPRec == nil {
		fmt.Println("Error dispatching client message because no destination Peer Record was provided.")
		return nil
	}
	conn, err := connectTo(destPRec.Address, destPRec.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := pc.NewClientListeningServiceClient(conn)
	response, err := c.SendClientMessage(context.Background(), CliMsg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}
	return response
}

// DispatchEventMessage publishes event messages to the clients upon event happenings
func (n *Node) DispatchEventMessage(clientRecord *pc.PeerRecord, eventType pc.EventMessage_MessageType, fileOrLockName string) {
	conn, err := connectTo(clientRecord.Address, clientRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := pc.NewClientListeningServiceClient(conn)

	var msg pc.EventMessage

	switch eventType {
	case pc.EventMessage_FileContentModified:
		msg = pc.EventMessage{Type: pc.EventMessage_FileContentModified, FileName: fileOrLockName}
	case pc.EventMessage_LockAquisition:
		msg = pc.EventMessage{Type: pc.EventMessage_LockAquisition, LockName: fileOrLockName}
	case pc.EventMessage_ConflictingLock:
		msg = pc.EventMessage{Type: pc.EventMessage_ConflictingLock, LockName: fileOrLockName}
	case pc.EventMessage_MasterFailOver:
		msg = pc.EventMessage{Type: pc.EventMessage_MasterFailOver}
	}

	response, err := c.SendEventMessage(context.Background(), &msg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}

	if response.Type != pc.EventMessage_Ack {
		fmt.Println("Expecting event response", pc.EventMessage_Ack, "from event -", eventType.String(), "but got", response.Type)
	}
}

// Convenience Methods

// publishMessage sends  messages to all subscribed clients - kinda like broadcast
func (n *Node) publishMessage(subscribed_clients []*pc.PeerRecord, eventType pc.EventMessage_MessageType, fileOrLockName string) {
	for _, clientRecord := range subscribed_clients {
		fmt.Println("TEST sending client", clientRecord.Address, clientRecord.Port, ": Event Type", eventType.String(), fileOrLockName)
		n.DispatchEventMessage(clientRecord, eventType, fileOrLockName)
	}
}
