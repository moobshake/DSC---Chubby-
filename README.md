# Bully Algorithm

This corresponds to the second half of Assignment 1.

## Pre-Requisites

Please install Protocol Buffers 3 for use with GoLang. It can be obtained from [here](https://developers.google.com/protocol-buffers/docs/downloads).

After installation, please update your system's environment variables or place the protoc.exe file into your GoPATH. Additionally, please also run the two following commands to downlooad the protoc plugins that are required for protoc to work with GoLang.  

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
```
This external package is required as this implemenation of the Bully Algorithm uses actual networking instead of using goroutines to simulate a client-server architecture.

<br>
<br>

## General Usage - Some Important Commands

### To run the BullyNode
```bash
go run . -id=ID -idOfMaster=ID -ipAddr="X.X.X.X" -port="0000" -verbose=OPTION
```
1) id: Specifies the ID of the node.
2) idOfMaster: Specifies the ID of the master node.
3) ipAddr: Specifies the IPv4 address of the node.
4) port: Specifies the listening port of the node.
5) verbose: Set to 1 to disable verbose logging and set to 2 to enable verbose logging.

Note: Please do not set the ID of a node to anything less than 1, and especially not 0. This limitation is because I have not properly handled some behaviours of protocol message serialisation. This is further explained in the Design Choices section below.

### To add a PeerRecord (information about a peer)
```bash
addPeer ID:IPv4_ADDRESS:Port ID:IPv4_ADDRESS:Port ...
```
Note: It is important to add the PeerRecord of the master Node if the node is not the master node. It is also unnecessary and impossible to add a peer record for the node itself.

### To online the node
```bash
online
```
This instructs the node to create or join a network, depending on whether the node is configured to be a master or to look for a master.

### To send a message to another node
```bash
msg ID MESSAGE
```
### To manually trigger an election
```bash
startElection
```
This command spoofs a Election message to the current node to trigger it into starting an election.

<br>
<br>

## For Checkoff
### Set-up
Run 5 nodes as such. The following 5 commands should be ran in 5 different terminals.
```bash
go run . -id=1 -idOfMaster=1 -ipAddr="127.0.0.1" -port="9090" -verbose=1
go run . -id=2 -idOfMaster=1 -ipAddr="127.0.0.1" -port="9091" -verbose=1
go run . -id=3 -idOfMaster=1 -ipAddr="127.0.0.1" -port="9092" -verbose=1
go run . -id=4 -idOfMaster=1 -ipAddr="127.0.0.1" -port="9093" -verbose=1
go run . -id=5 -idOfMaster=1 -ipAddr="127.0.0.1" -port="9094" -verbose=1
```
Online the network starting at the master node (1).
```bash
online
```
Online the rest of the network by running the following commands on the other nodes.
```bash
addPeer 1:127.0.0.1:9090
online
```
If desired, verify the status of the nodes and network with the two following commands.
```bash
getStatus
getPeers
```
Trigger an election so that the master node automatically becomes the node with the highest id (5). Run the following command on any node. Note: The election may take a few seconds to complete. The ID of the master node will be printed when the election is complete.
```bash
startElection
```

When the above commands are ran successfully, there should be a Bully Network with 5 nodes. 

Note: It is possible to skip triggering the election by starting modifying the go run -idOfMaster flag such that it is equal to the node with the higest ID.

IMPORTANT: The following instructions in each task assume the above setup. It should generally be possible to restore the setup state without repeating all the above instructions by simpling onlining whatever nodes were disabled and running an election with startElection.

<br>
<br>

### Task 1a: Simulate the worst-case election scenario
The worst case scenario is if the node with the lowest ID discovers that the master node is down and triggers an election.

To simluate this, have the master node leave the network in an ungraceful manner. This can be accomplished by CTRL+C or by running
```bash
exit
```
Note: If the "offline" command is used instead, the master node will gracefully exit the network by appointing the next highest ID node as the coordinator. 

Now, with the master node offline, have the node with the lowest ID attempt to send a message to the master node.
```bash
msg 5 Hello, are you online?
```
This will cause the node on which this command is issued to detect that the master node is down. When it confirms this with a KeepAlive message, it will start an election by attempting to elect itself.

Note: It is possible to examine the messages being passed from node to node during the election by turning verbose mode on (setting verbose=2). The message types will be displayed as numbers in the terminal, and they corresponding to constants in NodeCommFlags.go. For convenience: Type 1 = ElectionResults, Type 2 = ElectSelf, and Type 3 = RejectElect. With reference to the Bully Algorithm, the ElectSelf messages are sent by a node to higher ID nodes declaring its intention to elect itself as the new coordinator. RejectElect messages are sent by nodes to lower ID nodes upon receipt of an ElectSelf message to instruct said nodes to cease their self-election. ElectionResults messages are sent by the node that becomes the new coordinator to inform the network of its victory and of the new state of the network.

<br>
<br>

### Task 1b: Simulate the best-case election scenario
The best case scenario is if the node with the second highest ID after the failed failed cooridinator node discovers that said node is down and triggers an election.

Again, to simluate this, have the master node leave the network in an ungraceful manner. This can be accomplished by CTRL+C or by running
```bash
exit
```
Now, with the master node offline, have the node with the second highest ID send a message to the master node.
```bash
msg 5 Hello, are you online?
```

<br>
<br>

### Task 2a: Have the newly elected coordinator fail while announcing that it has won the election
In such a scenario, the end result of the network is that some of the nodes will have received the election results and others will not. Additionally, the newly elected coordinator will be offline.

Before moving on, I wish to state that due to certain implementation choices I have made, it is extremely difficult, if not impossible, for this scenario to even occur. This is because each node keeps track of whether there is an ongoing election. Each node only accepts that an election is finished when they receive a election result from a newly appointed coordinator. Consequently, if they do not receive such a election result message after a certain timeout period, they will automatically re-trigger an election under the assumption that the node that sent them the RejectElect might have gone offline during the election. Furthermore, when a node wins an election, it also makes the assumption that all higher ID node are offline. Consequently, it will remove the PeerRecords of those higher ID nodes from the network.

Nevertheless, it is still possible to force the network into such a fractured state as is desired by this task through some manual configurations.

Offline the coordinator node (5). On any two nodes (3 and 4), set their idOfMaster to node 4. Leave the other two nodes (1 and 2) untouched.
```bash
config idOfMaster=4
```
This simulates a situation of the original bully algorithm where node 4 won the election and node 3 received the result but nodes 1 and 2 did not.

This state of the network will eventually be rectified when any election occurs, and in this state, two ways by which an election by triggered are:
1) Any node that did not receive the election result (1 and 2) detects that the newly elected coordinator is down and initiates an election.
2) The newly elected coordinator (4) attempts to send a coordinator-only message to the nodes that did not receive an election (1 and 2). These nodes will reply to the coordinator with "NotMaster". This alerts the coordinator that the network may be fractured and so it initiates an election.

Scenario 1 can be simulated by having a node that did not receive the result attempt to message the downed master node.

Scenario 2 can be simulated by having a node that received the election result (3) go "offline", as this will prompt the coordinator (4) to update the rest of the network. As this update is a coordinator-only action, it will receive "NotMaster" replies from nodes that do not know that it is the coordinator (1 and 2) and initate an election. Scenario 2 also means that even if the original coordinator that failed re-awakens the network will still eventually converge.

```bash
offline
```

Note: I would like to reiterate that this state of the network should be impossible to occur as the nodes are implemented in such a way so as to re-trigger an election if they do not receive an election result in a timely manner.


<br>
<br>

### Task 2b: Have a node that will not win an election fail during the election
This scenario can be simulated by first triggering an election from any node, then ungracefully shutting down a node that will not win the election.

Run this on any node.
```bash
startElection
```

Then CTRL+C or run the following on a node that will not win the election.
```bash
shutdown
```

Due to the fact that the GRPC library uses HTTP/2 over TCP, a newly elected coordinator will rapidly discover the offline status our downed node when it attempts to broadcast the election results. Consequently, the coordinator will update the network about the status of our downed node.

Aside from this however, this scenario is also insignificant to the integrity of the network. Say that, for some reason, the newly appointed coordinator does not detect the downed node when it announces its election results, then when any node attempts to communicate with the downed node one of two things will happen.

1) If the communicating node is not the coordinator, it will notify the coordinator that our node is potentially offline. The coordinator will then verify this and update the network accordingly.
2) If the communciating node is the coordinator, it will verify the status of the node and update the network accordingly.

The two above possibilities can be inspected by simply ungracefully shutting down any node that is not the coordinator after an election before attempting to communicate with it from any node.
```bash
msg ID_OF_DOWNED_NODE Hello, are you online?
```

Either way, the state of the network will eventually become correct.
<br>
<br>

### Task 3: Multiple nodes start the election process simultaneously.
This is a very possible scenario and this can be simulated by offlining the coordinator and having more than one node attempt to message the coordinator at nearly the same time OR by simply having multiple nodes run the startElection commmand.

For the former method, press CTRL+C or run the following command on the coordinator node.
```bash
exit
```
Now, send a message to the offline coordinator from at multiple nodes within a second of one another.
```bash
msg 4 Hello, are you online?
```

For the latter method, rapidly run the following command on multiple nodes.\
```bash
startElection
```

This scenario will not cause any difficulty for this implementation of Bully Algorithm as nodes are able to keep track of whether there is an ongoing election. If an election is already ongoing, receiving an electSelf that may or may not be caused by the start of another election will simply cause the node to treat it as part of the first election.

<br>
<br>

### Task 4: An arbitrary and fresh node joins the network while no election is taking place.
Due to certain design choices in the implementation and as this implementation uses actual networking instead of simulating a client-server situation using goroutines, this task was already achieved in the set-up.

In this implementation of the bully algorithm, a new node only needs to know the IP_ADDRESS:PORT of any one node in an existing network for it to join that node. Then, when it attempts to join said network, a few scenarios can occur.
1) If the node the new node attempted to connect to is not the coordinator of its network, it sends a redirection notice to the coordinator to our new node that is automatically followed. 
2) If the node the new node attempted to connect to is the coordinator, the coordinator checks that there is no ID collision. If okay, the coordinator accepts the new node into the network. Following this, it updates the entire network of the new status of the network.

<br>
<br>

## Notes on some Design Choices
### 1) Redirection to Coordinator
This implementation of the Bully Algorithm contains is capable of providing redirection notices to the actual coordinator. This can most obviously be seen by having a new node attempt to join a network via a node that it thinks is the coordinator but is actually not. The node that the new node contacted will send a redirection to the actual coordinator of the network that will be automatically followed.

To see this, we can offline one node, set its idOfMaster to another node in the network, and have it attempt to online itself.
```bash
offline
config idOfMaster=ID_OF_SOME_OTHER_NODE_IN_THE_NETWORK
online
```

I implemented this as it makes the network more automatic and convenient. This allows me to have a node join an existing network so long as I know the IP address and port of any one node in the network.

### 2) Three kinds of Messages
This implementation of the Bully Algorithm involves NodeMessage, CoordinationMessage, and ControlMessage.
This allows me to logically categorise what kind of functions and messages should be sent to which RPC and with what sort of messages. Additionally, such a separation might allow different messages to be priotised for handling.

Note: ControlMessage was a last minute addition added only when I realised that GRPC duplicates every object when *grpc.Serve(lis net.Listener) is invoked. This was a huge source of frustration early on as I did not realise that my Node Structs, structures that logically represent the nodes, had been duplicated. This is because this behaviour necessarily implies a few design changes and experiments:
1) I cannot directly modify parameters in instances of Node from any goroutine because the goroutine running *grpc.Serve(lis net.Listener) uses attributes separate from every other goroutine.
2) I also attempted to use the "unsafe" module in GO to obtain the memory address of my struct instances. The idea was to convert theses memory addresses into primitves, which the listening goroutine seemed to duplicate, and have said goroutine reconstruct the integers back into memory addresses. This did not work as memory addresses in GO can only be converted into unsafe integer pointers (uintptr) vice versa. The listener goroutine did not duplicate uintptrs. 
3) I eventually overcame this by having the client interface dispatch actual messages to the server, as though the client interface was a client. For this, I implemented these messages as ControlMessages.

### 3) Not using the default values of variables in GO
As it turns out, when GRPC serialies protocol messages, if a variable has the default value, it is not serialised at all. While I might have also been able to overcome this with some proper handling of nil values, I decided that it would be simpler to avoid using nil values entirely.

Consequently, if you view NodeCommFlags.go, you may see that I have DoNotUseZeroValX in the iota constants. Additionally, attributes that could have been booleans are integers that use the values 1 to denote false and 2 to denote true.

This is also why the ID of any node should not be set to 0.

### 4) The Structure of the Code
#### 1) NodeComm.go
Contains the Node struct, some functions related to its initialisation, onlining, offlining, and implementations of the RPC methods.

Acts as an "API" for NodeCommServer.go.
#### 2) NodeComm.proto
Contains the definition for various protocol messages as well as the function definitions of the RPC methods.
#### 3) NodeCommClient.go
Contains the implementations of client-side versions of the RPC methods and some related convenience methods. Eg: If NodeComm.go contains a SayHello(), NodeCommClient.go will contain DispatchSayHello(). 

Acts as an "API" for NodeCommCLI.go. 

Technically, every supported interaction is possible with the methods in this file, but it is more convenient for a UI to be built on-top of this file.
#### 4) NodeCommServer.go
Contains methods that utilise the RPC methods implemented in NodeComm.go. Code in this file are strictly for the listener server, and should only be called by the listener. 
#### 5) NodeCommFlags.go
Contains constants/enums for use as flags in NodeMessage, ControlMessage, and CoordinationMessage.
#### 6) NodeCommCLI.go
A command line interface implemented using the methods in NodeCommClient.go. 
#### 7) main.go
The driver.
#### 8) NodeComm_grpc.pb.go
An automatically generated file from protocol buffers generated from NodeComm.proto.
#### 9) NodeComm.pb.go
An automatically generated file from protocol buffers generated from NodeComm.proto.

<br>
<br>

## Open Issues
1) The code is rather messy and in need of re-factoring and tidying up.
2) The protocol message serialisation behaviour of not serialising variables with default values needs to be properly handled.
3) Not all possible states of the node struct have been properly handled so the program may unexpectedly halt if an invalid configuration is manually set.