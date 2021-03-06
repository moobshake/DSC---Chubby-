package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	GRPC_CLIENT_TO_SEVER_CONN_TIMEOUT = 1 * time.Second
)

//connectTo abstracts three lines of grpc client code
func connectTo(address, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err

	// Putting this here in case we want a timeout
	// ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	// conn, err := grpc.DialContext(ctx, address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	fmt.Println("Dial Error:", err)
	// }
	// return conn, err
}

//  Table Methods
// Represents the data structure in the lookup table
// Fields are capitalized as only exported fields of a GO struct will be present in a JSON output
type lookupVal struct {
	IP   string
	Port string
}

// readLookup converts lookup json into an array of addresses
func readLookup() []lookupVal {
	lookupT, err := os.Open(lookup_path)

	if err != nil {
		fmt.Println(err)
	}

	defer lookupT.Close()
	byteValue, _ := ioutil.ReadAll(lookupT)
	var result []lookupVal
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

//update_Lookup is a method to add addresses to DNS table
func UpdateLookup(IP string, port string) {

	curTable := readLookup()

	newAddr := lookupVal{
		IP:   IP,
		Port: port,
	}

	curTable = append(curTable, newAddr)

	file, _ := json.MarshalIndent(curTable, "", "	")
	_ = ioutil.WriteFile(lookup_path, file, 0644)
}
