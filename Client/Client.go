package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//Create Client Struct here

// Client must be able to send request to DNS server to find master
// Client Struct inclusive of locks, id, peerrecord? Should we pass it in or have a global one?
//
const (
	dns_path = "lookup.json"
)

// Represents the data structure in the lookup table
// Fields are capitalized as only exported fields of a GO struct will be present in a JSON output
type lookup_v struct {
	NodeID int32
	Ip     string
	Port   string
}

// Client features
type Client struct {
	ClientID   int
	Master_add lookup_v
	Lock       int
	Handle     int
	Action     int
}

// Methods to implement
// 1. Looking for master via master location request
// 2. Directs any request it has to the specified address
// 3. Ways to edit the lookup table? (Done by a simple replacement system)

// Initialises a client
// Notes: Assuming 0 is this state where it has nothing
func CreateClient(id int, lock int, Master_add lookup_v, Action int) *Client {

	c := Client{
		ClientID: id,
		Lock:     0,
		Handle:   0,
		Action:   0,
	}

	return &c
}

// Master location request
func FindMaster() lookup_v {
	dnsTable := read_DNS()
	
}

// DNS Table Methods

// Converts lookup json into an array of addresses
func read_DNS() []lookup_v {
	lookupT, err := os.Open(dns_path)

	if err != nil {
		fmt.Println(err)
	}

	defer lookupT.Close()
	byteValue, _ := ioutil.ReadAll(lookupT)
	var result []lookup_v
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

// Method to add addresses to DNS table
func update_DNS(nodeID int32, IP string, port string) {

	curDNS := read_DNS()

	new_addr := lookup_v{
		NodeID: nodeID,
		Ip:     IP,
		Port:   port,
	}

	curDNS = append(curDNS, new_addr)

	file, _ := json.MarshalIndent(curDNS, "", "	")
	_ = ioutil.WriteFile(dns_path, file, 0644)
}

func main() {

}
