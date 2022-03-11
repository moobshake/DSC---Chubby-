package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	NC "assignment1/main/NodeComm"
)

//Create Client Struct here
const (
	lookup_path = "Client/lookup.json"
)

// Represents the data structure in the lookup table
// Fields are capitalized as only exported fields of a GO struct will be present in a JSON output
type lookup_val struct {
	IP   string
	Port string
}

// Client struct
type Client struct {
	ClientID  int
	ClientAdd *lookup_val
	MasterAdd *NC.PeerRecord
	Lock      *NC.Lock
	Action    int
}

// Methods to implement
// 1 - Master Location Request
// 2 - Read Message Request
// 3 - Write Message Request

// Initialises a client
// Notes: Majority of them are placeholder values
func CreateClient(id int, ipAdd, port string) *Client {

	c := Client{
		ClientID:  id,
		ClientAdd: &lookup_val{IP: ipAdd, Port: port},
		MasterAdd: &NC.PeerRecord{},
		Lock:      &NC.Lock{},
	}

	return &c
}

// Starting the client
func (c *Client) StartClient() {
	// go c.clientListener()
	c.FindMaster()
	fmt.Println("Client has a master now.")
	time.Sleep(1 * time.Second)
	c.startCLI()
}

// Master location request
func (c *Client) FindMaster() {
	lookupTable := read_Lookup()

	// Sending request for master node to every address listed in lookup json
	for i := 0; i < len(lookupTable); i++ {
		loc := lookupTable[i]
		pr := NC.PeerRecord{
			Id:      int32(-1),
			Address: loc.IP,
			Port:    loc.Port,
		}
		cm := NC.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     NC.ClientMessage_FindMaster,
			Spare:    -1,
			Message:  -1,
		}

		fmt.Printf("Client %d sent master request to: %s:%s\n", c.ClientID, loc.IP, loc.Port)
		res := c.DispatchClientMessage(&pr, &cm)
		fmt.Println(res)
	}

	masterIP := "127.0.0.1"
	masterPort := "9090"
	// Hardcoded Master address temporarily
	c.MasterAdd.Address = masterIP
	c.MasterAdd.Port = masterPort
	fmt.Printf("Master Node Registered: %s:%s\n", masterIP, masterPort)
}

// Making request
// Types - Write, Read
func (c Client) ClientRequest(reqType string) {

	var cm NC.ClientMessage

	switch reqType {
	case "Read":
		cm = NC.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     NC.ClientMessage_FileRead,
		}
		fmt.Printf("Client %d creating Read Request\n", c.ClientID)

	case "Write":
		cm = NC.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     NC.ClientMessage_FileWrite,
		}
		fmt.Printf("Client %d creating Write Request\n", c.ClientID)
	}

	res := c.DispatchClientMessage(c.MasterAdd, &cm)

	// fmt.Printf("Master replied: %d, Message: %d\n", res.Type, res.Message)
	fmt.Printf("Master replied: %d", res.Type)
}

// Lookup Table Methods
// Converts lookup json into an array of addresses
func read_Lookup() []lookup_val {
	lookupT, err := os.Open(lookup_path)

	if err != nil {
		fmt.Println(err)
	}

	defer lookupT.Close()
	byteValue, _ := ioutil.ReadAll(lookupT)
	var result []lookup_val
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

// Method to add addresses to DNS table
func update_Lookup(nodeID int32, IP string, port string) {

	curTable := read_Lookup()

	new_addr := lookup_val{
		IP:   IP,
		Port: port,
	}

	curTable = append(curTable, new_addr)

	file, _ := json.MarshalIndent(curTable, "", "	")
	_ = ioutil.WriteFile(lookup_path, file, 0644)
}
