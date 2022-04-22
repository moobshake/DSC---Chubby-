# Our Chubby Algorithm ^_^

## General Usage
When first ran, our program starts an initial CLI interface that can be used to create, load, modify, and save start parameters. Upon "starting", you will be brought to either the CLI of the client or the CLI of a chubby replica.

## Initial CLI
### Overview of Available Commands
```
Available Commands:
'help':                         Prints this help message or information about a given command.
'config':                       Used to set various parameters.
'guided-config':                Starts guided configuration.
'start':                        Starts the node.
'saveConfig':                   Saves the current configuration to file.
'listConfigs':                  Lists all known configurations.
'loadConfig':                   Loads a specified configuration from file.
'showConfig':                   Prints the current configuration.
'printConfig':                  An alias for 'showConfig'.
'startConfig':                  Loads and starts a specified configuration.
'updateDNS [IP] [PORT]':        Allows you to update the DNS table'.
'exit':                         Exits the program.
```

### help
This command prints out the list of avaialble commands.
```
> help <command>
```
Additionally, if the name of a command is provided as an argument, additional help text for the specified command will be provided if there is any.

### config
This command allows the parameters of the currently loaded configuration to be modified.
```
> config <param_name=param_value> <param_name=param_value> ...
```
Avaiable parameters are:
- nodeID: Any integer value in the range [0, inf)
- nodeIP: Any valid IPv4 address.
- nodePort: Any valid port number.
- verbose: 1 (off) and 2 (on).
- idOfMaster: Node ID of master server.
- configName: A name for this configuration. Whitespaces are not supported.
- nodeType: 's' or 'server' for Server and 'c' or 'client' for client.

Example Usage:
```
> config nodeID=11 nodeIP=127.0.0.1
```

### guided-config
This command launches a configuration wizard that walks the user through configuration.
```
> guided-config
```

### start
This command starts the current configuration. Assuming that the current configuration is valid, this command will also bring the user to either the CLI of the server or the CLI of the client, depending on the node type (nodeType) specified in the configuration.
```
> start
```

### saveConfig
This command will save the current configuration as a file.
```
> saveConfig <configuration_name>
```
Note: Whitespaces are not supported.

### listConfigs
This command will list all saved configurations.
```
> listConfigs
```

### loadConfig
This command will attempt to load the specified configuration from file.
```
> loadConfig <configuration_name>
```

### printConfig
This command prints the current configuration to the terminal.
```
> printConfig
```

### startConfig
This command is a convenience wrapper of loadConfig and start - it attempts to load then start the specified configuration.
```
> startConfig <configuration_name>
```

### udpateDNS
Yu Hui help please

### exit
This command exits the program.
```
> exit
```

## Client CLI
### Overview of Available Commands
```
Available Commands:
'help':                                          Prints this menu.
'exit':                                          Exit program.
'write FILE_NAME [modify file? true/false]':     Client sends Write Request to Master.
'read FILE_NAME':                                Client sends Read Request to Master.
'sub SUB_TYPE':                                  Sends a subscription request. Type help sub for more info.
'ls':                                            Lists the files available for the client
'll':                                            Lists the locks available for the client
'releaseLock FILE_NAME':                         Release lock for FILE_NAME
'requestLock LOCK_TYPE FILE_NAME':               Requests lock with LOCK_TYPE (read/write)for FILE_NAME
```

### help

### exit

### write FILE_NAME

### read FILE_NAME

### sub SUB_TYPE

### ls

### ll

### releaseLock

### requestLock

## Server CLI
### Overview of Available Commands
```
Available Commands:
'exit':                         Exit program.
'online':                       Onlines the node.
'offline':                      Offlines the node.
'config':                       Configuring node parameters.
'addPeer':                      Adds a record about a peer node.
'delPeer':                      Deletes a peer node record.
'getPeer':                      Get the list of peer records.
'msg':                          Send a plain text message to a peer node.
'getStatus':                    Obtain information about the node.
'startElection':                Used to manually trigger an election.
'publish':                      Used to manually publish an update to subscribed clients.
'wakeUpNode [ADDR:PORT]':       Used to manually wake up another replica.
'help':                         Prints this menu.
```

### exit
This command terminates the server and returns the user to the initial CLI.
```
> exit
```
Note: This command may not entirely terminate the server for quite a whie. To properly simulate a failed replica, we recommend terminating the entire program.

### online
This command 'onlines' the server replica. In this state, the replica will respond to messages.
```
> online
```

### offline
This command 'offlines' the server replica. In this state, the replica will ignore all messages except remote wake up messages from a coordinator.
```
> offline
```
Note: This command does not work correctly. To properly simulate a failed replica, we recommend terminating the entire program.

### config
This command modifies certain parameters of the server replica.
```
> config <param_name=param_value> <param_name=param_value> ...
```
Available parameters:
- id              =       ID_OF_THIS_NODE
- idOfMaster      =       ID_OF_MASTER
- verbose         =       1 or 2 for off and on respectively

### addPeer
This command addes a peer record to the server replica.
```
> addPeer NODE_ID:NODE_IP:NODE_PORT
```
Example:
```
> addPeer 2:127.0.0.1:9091
```

### delPeer
This command deletes the specified peer record from the server replica.
```
> delPeer NODE_ID
```

### getPeers
This command prints all peer records known to the server replica.
```
> getPeers
```

### msg
This command sends a single text message to another replica.
```
> msg <TEXT...>
```

### getStatus
This command prints some information about the replica.
```
> getStatus
```

### startElection
This command instructs the replica to initate an election within the replica network.
```
> startElection
```

### publish
Hannah help

### wakeUpNode
This command is used by the coordinator to remotely instruct an offline replica to online and join the coordinator's network.
```
> wakeUpNode NODE_IP:NODE_PORT
```
Note: The joining node will attempt to use its configured ID. If another node in the network already has the same ID however, its ID will be automatically changed.

### help
This command prints the help menu to the terminal.
```
> help
```

## Code Structure
![](code_structure.jpg)

## Protocol Buffer Generation
```
protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative  ./ProtocChubby/Chubby.proto
```
