package nodecomm

import (
	pc "assignment1/main/protocchubby"
	"context"
	"fmt"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

// Write requests should be acknowledged by the majority of replicas
// SendWriteRequestToReplicasUtil sends a stream of messages from the master to the replicas.
// The replicas should save the file to their local copies and return an OK if they are done
// Otherwise, send back an error.
// This function cannot be broken up further without overcomplicating things.
func (n *Node) DispatchWriteRequestToReplicasUtil(writeMsgBuffer []*pc.ClientMessage, peerRecord *pc.PeerRecord, replyChan chan bool) {
	// create stream for message sending
	conn, err := connectTo(peerRecord.Address, peerRecord.Port)
	if err != nil {
		replyChan <- false
		return
	}

	defer conn.Close()

	cli := pc.NewNodeCommListeningServiceClient(conn)

	stream, err := cli.SendWriteRequest(context.Background())
	if err != nil {
		replyChan <- false
		return
	}

	for _, cliWriteMsg := range writeMsgBuffer {
		// The file content to embed in the request
		cliWriteMsg.Type = pc.ClientMessage_ReplicaWrites

		if err := stream.Send(cliWriteMsg); err != nil {
			fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println("MASTER FILE WRITE REQUEST TO REPLICA ERROR:", err)
		replyChan <- false
		return
	}

	replyChan <- reply.Type == pc.ClientMessage_Ack

	fmt.Println("Master write to replica reply from replica:", reply.Type)
}
