package nodecomm

//Flags here are only to be used in NodeMessages
type NodeMsgType int32

const (
	DoNotUseZeroVal1 NodeMsgType = iota
	KeepAlive
	Standard
)

//Flags here are only to be used in CoordinationMessages
type CoordinationMsgType int32

const (
	DoNotUseZeroVal2 CoordinationMsgType = iota
	ElectionResult
	ElectSelf
	RejectElect
	ReqToJoin
	RejectJoin
	PeerInformation
	RedirectToCoordinator
	ReqToMerge
	RejectMerge
	ApptNewCoordinator
	ReqToLeave
	BadNodeReport
	NotMaster
)

//Flags here are only to be used in ControlMessages
type ControlFlag int32

const (
	DoNotUseZeroVal3 ControlFlag = iota
	StopListening
	InitParams
	UpdateParams
	GetStatus
	GetPeers
	AddPeer
	DelPeer
	Okay
	Message
	RawMessage
	JoinNetwork
	LeaveNetwork
	StartElection
	GetParams
)

type GenericMessageFlags int32

const (
	DoNotUseZeroVal4 GenericMessageFlags = iota
	Empty
	Ack
	Warning
	Error
	NotInNetwork
)

type ClientMessageFlags int32

const (
	DoNotUserZeroVal5 ClientMessageFlags = iota
	FindMaster
	FileRead
	FileWrite
)
