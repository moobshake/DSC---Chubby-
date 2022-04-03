package client

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"time"

	pc "assignment1/main/protocchubby"

	"google.golang.org/grpc"
)

const (
	lookup_path = "Client/lookup.json"
	// Full cache file location:
	//     CACHE_ROOT/CACHE_DIR_PREFIX_CLIENT_ID/file
	CACHE_ROOT       = "client_cache"
	CACHE_DIR_PREFIX = "client_cache"
)

// Client struct
// Lock is a string temporarily (Assume acquiring one lock)
type Client struct {
	pc.ClientControlServiceServer
	pc.ClientListeningServiceServer
	ClientID  int
	ClientAdd *lookupVal
	MasterAdd *pc.PeerRecord
	Locks     map[string]lock // Map where key = filename, value = lock details
	Action    int
	// This is where the full file path to the client's cache
	ClientCacheFilePath string

	// Check if the file in cache is valid
	// Valid means that to the client's knowledge,
	// the cached file is the same as the master's
	// The client will know if the file is invalid if
	// the client subscribes to a master event.
	// Key: file name, val: validity of file
	ClientCacheValidation map[string]bool
}

// Methods to implement
// 1 - Master Location Request
// 2 - Read Message Request
// 3 - Write Message Request

// CreateClient initialises a client
// Notes: Majority of them are placeholder values
func CreateClient(id int, ipAdd, port string) *Client {

	c := Client{
		ClientID:              id,
		ClientAdd:             &lookupVal{IP: ipAdd, Port: port},
		MasterAdd:             &pc.PeerRecord{},
		Locks:                 map[string]lock{},
		ClientCacheFilePath:   filepath.Join(CACHE_ROOT, CACHE_DIR_PREFIX+"_"+strconv.Itoa(id)),
		ClientCacheValidation: make(map[string]bool),
	}

	return &c
}

// StartClient starts the client
func (c *Client) StartClient() {
	if !c.FindMaster() {
		return
	}
	go c.startClientListener()
	time.Sleep(1 * time.Second)
	c.startCLI()
}

//startClientListener starts a *grpc.serve server as a listener.
func (c *Client) startClientListener() {
	fullAddress := c.ClientAdd.IP + ":" + c.ClientAdd.Port
	fmt.Println("Starting listener at ", fullAddress)
	lis, err := net.Listen("tcp", fullAddress)
	if err != nil {
		log.Fatalf("Failed to hook into: %s. %v", fullAddress, err)
	}

	clistener := Client{
		ClientID:              c.ClientID,
		ClientAdd:             &lookupVal{IP: c.ClientAdd.IP, Port: c.ClientAdd.Port},
		MasterAdd:             c.MasterAdd,
		Locks:                 map[string]lock{},
		ClientCacheFilePath:   c.ClientCacheFilePath,
		ClientCacheValidation: make(map[string]bool),
	}

	gServer := grpc.NewServer()
	pc.RegisterClientControlServiceServer(gServer, &clistener)
	pc.RegisterClientListeningServiceServer(gServer, &clistener)

	if err := gServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}

// FindMaster sends a master location request
func (c *Client) FindMaster() bool {
	lookupTable := readLookup()

	connected := false
	// Sending request for master node to every address listed in lookup json
	for i := 0; i < len(lookupTable); i++ {
		loc := lookupTable[i]
		pr := pc.PeerRecord{
			Id:      int32(-1),
			Address: loc.IP,
			Port:    loc.Port,
		}
		cm := pc.ClientMessage{
			ClientID: int32(c.ClientID),
			Type:     pc.ClientMessage_FindMaster,
			Spare:    -1,
			Message:  -1,
		}

		fmt.Printf("Client %d sent master request to: %s:%s\n", c.ClientID, loc.IP, loc.Port)
		res := c.DispatchClientMessage(&pr, &cm)
		fmt.Println(res)
		if res == nil {
			fmt.Println("Client failed to connect to Chubby replica", i)
		} else if res.Type == pc.ClientMessage_ConfirmCoordinator {
			c.MasterAdd.Address = loc.IP
			c.MasterAdd.Port = loc.Port
			fmt.Printf("Master Node Registered: %s:%s\n", loc.IP, loc.Port)
			connected = true
			return connected
		} else if res.Type == pc.ClientMessage_RedirectToCoordinator {
			c.HandleMasterRediction(res)
			connected = true
			return connected
		}
	}
	fmt.Println("Client is unable to connect to any of its known masters. Exiting...")
	return false
}

func (c *Client) DummyFindMaster() bool {
	masterIP := "127.0.0.1"
	masterPort := "9090"
	// Hardcoded Master address temporarily
	c.MasterAdd.Address = masterIP
	c.MasterAdd.Port = masterPort
	fmt.Printf("Master Node Registered: %s:%s\n", masterIP, masterPort)
	return true
}

func (c *Client) HandleMasterRediction(redirectionMsg *pc.ClientMessage) {
	c.MasterAdd.Address = redirectionMsg.ClientAddress.Address
	c.MasterAdd.Port = redirectionMsg.ClientAddress.Port
	fmt.Printf("Master Node Registered: %s:%s\n", c.MasterAdd.Address, c.MasterAdd.Port)
}
