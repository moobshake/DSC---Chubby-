package nodecomm

import (
	"context"
	"fmt"

	pc "assignment1/main/protocchubby"
)

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><> Dispatch Methods - These methods SEND messages <><><>

//DispatchShutdown sends a shutdown signal to the server
func (n *Node) DispatchShutdown() bool {
	conn, err := connectTo(n.myPRecord.Address, n.myPRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := pc.NewNodeCommControlServiceClient(conn)

	cMsg := pc.ControlMessage{Type: pc.ControlMessage_StopListening}

	response, err := c.Shutdown(context.Background(), &cMsg)
	if err != nil {
		fmt.Println("Error dispatching shutdown signal:", err)
	}
	if response.Type == pc.ControlMessage_Okay {
		return true
	}
	return false
}

//DispatchControlMessage dispatches a control message to the server
func (n *Node) DispatchControlMessage(cMsg *pc.ControlMessage) *pc.ControlMessage {
	conn, err := connectTo(n.myPRecord.Address, n.myPRecord.Port)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	c := pc.NewNodeCommControlServiceClient(conn)
	response, err := c.SendControlMessage(context.Background(), cMsg)
	if err != nil {
		fmt.Println("Error dispatching control message:", err)
	}
	return response
}
