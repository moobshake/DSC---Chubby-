package nodecomm

import (
	"context"
	"fmt"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> RPC Send Methods - These methods RECEIVE messages <><><>

//KeepAlive - Used to check if a node is alive or not.
func (n *Node) KeepAlive(ctx context.Context, inMsg *pc.NodeMessage) (*pc.NodeMessage, error) {
	if !n.isOnline {
		return &pc.NodeMessage{Type: pc.NodeMessage_Empty}, nil
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		return &pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_NotInNetwork}, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case pc.NodeMessage_KeepAlive:
		nMsg := pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_KeepAlive, Comment: inMsg.Comment}
		return &nMsg, nil
	default:
		//Received a non-keepalive message on the keepalive channel
		return &pc.NodeMessage{Type: pc.NodeMessage_Error, Comment: "Unsupported function."}, nil
	}
}

//SendMessage is a channel for pc.NodeMessages
func (n *Node) SendMessage(ctx context.Context, inMsg *pc.NodeMessage) (*pc.NodeMessage, error) {
	if !n.isOnline {
		return &pc.NodeMessage{Type: pc.NodeMessage_Empty}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Message:", inMsg)
	}
	if !n.isPeer(inMsg.FromPRecord.Id) {
		nMsg := pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_MessageType(pc.ClientMessage_RedirectToCoordinator)}
		nMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
		nMsg.Spare = int32(n.idOfMaster)
		return &nMsg, nil
	}
	n.updatePeerRecords(inMsg.FromPRecord)
	switch inMsg.Type {
	case pc.NodeMessage_Standard:
		fmt.Println(inMsg.FromPRecord.Id, ": ", inMsg.StandardBody.Message)
		return &pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_Ack}, nil
	default:
		//Received a non-standard message on the standard channel
		//Consider returning an error or doing something else
		return &pc.NodeMessage{FromPRecord: n.myPRecord, Type: pc.NodeMessage_Warning, Comment: "Message sent to wrong channel."}, nil
	}
}

//SendCoordinationMessage is a channel for Coordination Messages
func (n *Node) SendCoordinationMessage(ctx context.Context, coMsg *pc.CoordinationMessage) (*pc.CoordinationMessage, error) {
	if !n.isOnline {
		if coMsg.Type == pc.CoordinationMessage_WakeUpAndJoinNetwork {
			fmt.Println("Recieved:", coMsg.Type)
			// if n.idOfMaster != int(coMsg.FromPRecord.Id) {
			// 	return &pc.CoordinationMessage{Type: pc.CoordinationMessage_NotMaster}, nil
			// }
			fmt.Println("This node has received a WAKEUP message from a coordinator.")
			n.idOfMaster = int(coMsg.FromPRecord.Id)
			n.updatePeerRecords(coMsg.FromPRecord)
			//Quicky and dirty ID collision avoidance
			for _, pRec := range n.peerRecords {
				if pRec.Id == n.myPRecord.Id {
					n.myPRecord.Id = int32(len(n.peerRecords) + 1)
					break
				}
			}
			n.onlineNode()
			response := n.DispatchCoordinationMessage(coMsg.FromPRecord, &pc.CoordinationMessage{Type: pc.CoordinationMessage_ReqToMirror})
			if response.Type == pc.CoordinationMessage_Ack {
				return &pc.CoordinationMessage{FromPRecord: n.myPRecord, Type: pc.CoordinationMessage_Ack}, nil
			} else {
				n.offlineNode()
				return &pc.CoordinationMessage{Type: pc.CoordinationMessage_Empty}, nil
			}
		}
		return &pc.CoordinationMessage{Type: pc.CoordinationMessage_Empty}, nil
	}
	if n.verbose == 2 {
		fmt.Println("Received Coordination Message:", coMsg)
	}
	nCoMsg := pc.CoordinationMessage{FromPRecord: n.myPRecord, Type: pc.CoordinationMessage_Error}
	//Node is NOT the master and it received an out of network message
	if !n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case pc.CoordinationMessage_ReqToJoin: //Send redirection
			coordinatorRecord := []*pc.PeerRecord{n.getPeerRecord(n.idOfMaster, false)}
			nCoMsg.Type = pc.CoordinationMessage_RedirectToCoordinator
			nCoMsg.PeerRecords = coordinatorRecord
		default:
			nCoMsg.Type = pc.CoordinationMessage_NotInNetwork
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is NOT the master and it received an in-network message
	} else if !n.IsMaster() {
		switch coMsg.Type {
		case pc.CoordinationMessage_ApptNewCoordinator: //Accept if message was from master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.idOfMaster = int(coMsg.Spare)
				nCoMsg.Type = pc.CoordinationMessage_Ack
				if n.IsMaster() {
					n.deletePeerRecord(int(coMsg.FromPRecord.Id))
					n.BroadcastPeerInformation()
				}
			} else {
				nCoMsg.Type = pc.CoordinationMessage_NotMaster
				nCoMsg.Spare = int32(n.idOfMaster)
			}
		case pc.CoordinationMessage_MirrorRecord:
			n.MirrorSink(coMsg.MirrorRecords)
		default:
			nCoMsg.Comment = "Unsupported function."
			nCoMsg.Type = pc.CoordinationMessage_Error
		}
		//Node is master and it received an out of network message
	} else if n.IsMaster() && !n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case pc.CoordinationMessage_ReqToJoin:
			pRec := n.getPeerRecord(int(coMsg.FromPRecord.Id), true)
			if pRec == nil { //Allow node to join, send peer records
				n.updatePeerRecords(coMsg.FromPRecord)
				nCoMsg.Type = pc.CoordinationMessage_PeerInformation
				nCoMsg.PeerRecords = append(n.peerRecords, n.myPRecord)
				//Broadcast info to all nodes
				n.BroadcastPeerInformation()
			} else { //ID collision, REJECT node
				nCoMsg.Type = pc.CoordinationMessage_RejectJoin
				nCoMsg.Comment = "ID Collision"
			}
		case pc.CoordinationMessage_ReqToMerge:
			nCoMsg.Comment = "Unimplemented funcion."
		default:
			nCoMsg.Type = pc.CoordinationMessage_NotInNetwork
			nCoMsg.Comment = "Sendee node is not part of the network. Send ReqToJoin to master node."
			nCoMsg.Spare = int32(n.idOfMaster)
		}
		//Node is master and it received an in-network message
	} else if n.IsMaster() {
		switch coMsg.Type {
		case pc.CoordinationMessage_ReqToLeave:
			nCoMsg.Type = pc.CoordinationMessage_Ack
			n.deletePeerRecord(int(coMsg.FromPRecord.Id))
			n.BroadcastPeerInformation()
		case pc.CoordinationMessage_BadNodeReport:
			if n.badNodeHandler(n.getPeerRecord(int(coMsg.Spare), false)) {
				//Master node could reach allegedly downed node. Send peer information to
				//node that made the report in case it was a case of bad network details.
				go n.DispatchCoordinationMessage(coMsg.FromPRecord, &pc.CoordinationMessage{Type: pc.CoordinationMessage_PeerInformation, PeerRecords: append(n.peerRecords, n.myPRecord)})
			}
			nCoMsg.Type = pc.CoordinationMessage_MessageType(pc.CoordinationMessage_Ack)
		case pc.CoordinationMessage_ReqToMirror:
			n.MirrorDispatcher(coMsg)
			nCoMsg.Type = pc.CoordinationMessage_Ack
		case pc.CoordinationMessage_ReqFile:
			MRec := coMsg.MirrorRecords[0]
			n.DispatchFileToReplica(coMsg.FromPRecord, MRec.FilePath)
		default:
			coMsg.Comment = "Unsupported operation."
		}
	}
	//Messages that apply for in-network nodes regardless of master status
	if n.isPeer(coMsg.FromPRecord.Id) {
		switch coMsg.Type {
		case pc.CoordinationMessage_PeerInformation: //Accept only if it came from the master
			if coMsg.FromPRecord.Id == int32(n.idOfMaster) {
				n.overridePeerRecords(coMsg.PeerRecords)
			} else if n.IsMaster() { //Some other node think its the master, fix by re-triggering election
				n.spoofElection()
			} else { //The sendee node should not be sending a master-only message
				nCoMsg.Comment = "Sender node is not coordinator."
				nCoMsg.Type = pc.CoordinationMessage_NotMaster
			}
		case pc.CoordinationMessage_ElectionResult, pc.CoordinationMessage_ElectSelf, pc.CoordinationMessage_RejectElect:
			n.electionHandler(coMsg)
		}
	}
	return &nCoMsg, nil
}

// EstablishReplicaConsensus is used to handle messages recieved from the master
// to the replicas. This is used to attempt to ensure consensus.
// Note: Write Requests are not processed here due to streaming
func (n *Node) EstablishReplicaConsensus(ctx context.Context, serverMsg *pc.ServerMessage) (*pc.ServerMessage, error) {
	fmt.Printf("Replica %d receiving message from master:%s\n", n.myPRecord.Id, serverMsg.Type)

	switch serverMsg.Type {
	// TODO: YH create the functions this case will use, feel free to change the serverMsg Type
	case pc.ServerMessage_ReqLock, pc.ServerMessage_ReadLock, pc.ServerMessage_WriteLock:
		// Find master
		// Replies with master's address
		// Function to handle saving locks by replica
		return n.handleLockfromMaster(serverMsg), nil

	case pc.ServerMessage_ReplicaReadCheck:
		return n.handleReadRequestFromMaster(serverMsg), nil

	case pc.ServerMessage_RelLock:
		return n.relLockfromMaster(serverMsg), nil

	case pc.ServerMessage_SubscribeFileModification, pc.ServerMessage_SubscribeLockConflict,
		pc.ServerMessage_SubscribeMasterFailover, pc.ServerMessage_SubscribeLockAquisition:
		n.ReplicaClientSubscriptionsHandler(serverMsg)
		return &pc.ServerMessage{Type: pc.ServerMessage_Ack}, nil
	default:
		fmt.Printf("> Replica requesting for something that is not available %s\n", serverMsg.Type)
	}

	return &pc.ServerMessage{Type: pc.ServerMessage_Error, StringMessages: "Request not available"}, nil
}

// Receive a stream of write messages from the master for replication.
func (n *Node) SendWriteForward(stream pc.NodeCommPeerService_SendWriteForwardServer) error {
	var writeRequestMessage *pc.ServerMessage
	writeRequestMessage, err := stream.Recv()
	if err != nil {
		return err
	}
	if !(writeRequestMessage.Type == pc.ServerMessage_ReplicaWriteData || writeRequestMessage.Type == pc.ServerMessage_ReplicaWriteLock) {
		return nil
	}
	return n.handleMasterToReplicatWriteRequest(stream, writeRequestMessage)
}
