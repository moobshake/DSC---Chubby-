package nodecomm

import (
	"context"
	"fmt"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> RPC Send Methods - These methods RECEIVE messages <><><>

//SendControlMessage is a Channel for Control Messages
func (n *Node) SendControlMessage(ctx context.Context, cMsg *pc.ControlMessage) (*pc.ControlMessage, error) {
	if n.verbose == 2 {
		fmt.Println("Received pc.ControlMessage:", cMsg)
	}
	switch cMsg.Type {
	case pc.ControlMessage_GetStatus:
		pBody := pc.ParamsBody{IdOfMaster: int32(n.idOfMaster), MyPRecord: n.myPRecord}
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay, ParamsBody: &pBody}, nil
	case pc.ControlMessage_GetPeers:
		pBody := pc.ParamsBody{PeerRecords: n.peerRecords}
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay, ParamsBody: &pBody}, nil
	case pc.ControlMessage_InitParams:
		n.electionStatus = &pc.ElectionStatus{OngoingElection: 1, IsWinning: 1, Active: 1, TimeoutDuration: int32(3)}
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		n.lockGenerationNumber = int(cMsg.ParamsBody.LockGenerationNumber)
		n.nodeDataPath = cMsg.ParamsBody.NodeDataPath
		n.nodeLockPath = cMsg.ParamsBody.NodeLockPath
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_GetParams:
		pBody := pc.ParamsBody{
			IdOfMaster: int32(n.idOfMaster),
			MyPRecord:  n.myPRecord,
			Verbose:    int32(n.verbose),
		}
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay, ParamsBody: &pBody}, nil
	case pc.ControlMessage_UpdateParams:
		n.idOfMaster = int(cMsg.ParamsBody.IdOfMaster)
		n.myPRecord = cMsg.ParamsBody.MyPRecord
		n.verbose = int(cMsg.ParamsBody.Verbose)
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_AddPeer:
		pRecords := cMsg.ParamsBody.PeerRecords
		n.mergePeerRecords(pRecords)
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_DelPeer:
		pRecords := cMsg.ParamsBody.PeerRecords
		for _, pRec := range pRecords {
			n.deletePeerRecord(int(pRec.Id))
		}
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_Message:
		nMessage := pc.NodeMessage{
			FromPRecord:  n.myPRecord,
			ToID:         cMsg.Spare,
			Type:         pc.NodeMessage_Standard,
			StandardBody: &pc.StandardBody{Message: cMsg.Comment}}
		n.DispatchMessage(n.getPeerRecord(int(cMsg.Spare), false), &nMessage)
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_RawMessage:
		n.DispatchMessage(n.getPeerRecord(int(cMsg.NodeMessage.ToID), false), cMsg.NodeMessage)
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_JoinNetwork:
		if !n.isOnline {
			n.onlineNode()
			return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
		}
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay, Comment: "Node is already online!"}, nil

	case pc.ControlMessage_LeaveNetwork:
		if n.isOnline {
			n.offlineNode()
			return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
		}
		return &pc.ControlMessage{Type: pc.ControlMessage_Error, Comment: "Node is already offline!"}, nil

	case pc.ControlMessage_StartElection:
		n.spoofElection()
		return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
	case pc.ControlMessage_PublishMasterFailover:
		n.PublishMasterFailover()
	case pc.ControlMessage_PublishFileModification:
		n.PublishFileContentModification(cMsg.Comment, nil)
	case pc.ControlMessage_PublishLockAquisition:
		n.PublishLockAcquisition(cMsg.Comment)
	case pc.ControlMessage_PublishLockConflict:
		n.PublishConflictingLockRequest(cMsg.Comment)
	}
	//TODO: Add here

	return &pc.ControlMessage{Type: pc.ControlMessage_Error, Comment: "Unsupported function."}, nil
}

//Shutdown - Unimplemented as I found that a graceful shutdown is unimportant for this.
func (n *Node) Shutdown(ctx context.Context, cMsg *pc.ControlMessage) (*pc.ControlMessage, error) {
	if cMsg.Type != pc.ControlMessage_StopListening {
		return &pc.ControlMessage{Type: pc.ControlMessage_Error, Comment: "Wrong channel."}, nil
	}
	return &pc.ControlMessage{Type: pc.ControlMessage_Okay}, nil
}
