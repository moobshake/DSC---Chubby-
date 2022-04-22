package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	Client "assignment1/main/Client"
	NodeComm "assignment1/main/NodeComm"
)

func main() {

	printHelp([]string{})
	config := &configBody{}
	scanner := bufio.NewScanner(os.Stdin)
	var userInput string
	fmt.Print("> ")

MainLoop:
	for scanner.Scan() {
		userInput = ""
		userInput = scanner.Text()
		tokenised := strings.Fields(userInput)

		if len(tokenised) == 0 {
			fmt.Print(">")
			continue
		}

		switch strings.ToLower(tokenised[0]) {

		case "config":
			manualConfig(&config, tokenised[1:], false)
		case "guided-config":
			guidedConfig(scanner, &config)
		case "start":
			startNode(config)
		case "saveconfig":
			saveConfig(config)
		case "listconfigs":
			listConfigs()
		case "loadconfig":
			loadConfig(&config, tokenised[1:])
		case "deleteconfig":
			deleteConfig(tokenised[1:])
		case "showconfig", "printconfig":
			printConfig(config)
		case "updatedns":
			Client.UpdateLookup(tokenised[1], tokenised[2])
			fmt.Println(tokenised[1], tokenised[2])
		case "startconfig":
			loadConfig(&config, tokenised[1:])
			startNode(config)
		case "help":
			printHelp(tokenised[1:])

		case "exit":
			break MainLoop

		default:
			fmt.Println("Invalid input.")
		}
		fmt.Print("> ")
	}
}

func manualConfig(currentConfig **configBody, args []string, silent bool) bool {
	if len(args) == 0 {
		fmt.Println("No arguments provided.")
		return false
	}

	newConfig := configBody(**currentConfig)
	errorDetected := false

mainConfigLoop:
	for _, arg := range args {
		tokenisedArgs := strings.Split(arg, "=")
		if len(tokenisedArgs) != 2 {
			fmt.Println(arg, "is not recognised.")
			errorDetected = true
			break mainConfigLoop
		}
		argName, argVal := tokenisedArgs[0], tokenisedArgs[1]
		switch argName {
		case "nodeID":
			tID, err := strconv.Atoi(argVal)
			if err != nil {
				fmt.Println(argVal, "is not a valid node id.")
				errorDetected = true
			} else {
				newConfig.NodeID = tID
			}
		case "nodeIP":
			octets := strings.Split(argVal, ".")
			if len(octets) != 4 {
				fmt.Println(argVal, "is not a valid IPv4 address.")
				errorDetected = true
				break mainConfigLoop
			}
			for _, octect := range octets {
				intOctect, err := strconv.Atoi(octect)
				if err != nil {
					errorDetected = true
					fmt.Println(argVal, "is not a valid IPv4 address.")
					break mainConfigLoop
				} else if intOctect < 0 || intOctect > 255 {
					errorDetected = true
					fmt.Println(argVal, "is not a valid IPv4 address.")
					break mainConfigLoop
				}
			}
			newConfig.NodeIP = argVal
		case "nodePort":
			tPort, err := strconv.Atoi(argVal)
			if err != nil {
				fmt.Println(tPort, "is not a valid port number.")
				errorDetected = true
				break mainConfigLoop
			} else if tPort < 0 || tPort > 65535 {
				fmt.Println(tPort, "is not a valid port number.")
				errorDetected = true
				break mainConfigLoop
			} else {
				newConfig.NodePort = argVal
			}
		case "verbose":
			tVerbose, err := strconv.Atoi(argVal)
			if err != nil {
				fmt.Println(tVerbose, "is not a valid verbosity setting.")
				errorDetected = true
				break mainConfigLoop
			} else if tVerbose < 0 || tVerbose > 1 {
				fmt.Println(tVerbose, "is not a valid verbosity setting.")
				errorDetected = true
				break mainConfigLoop
			} else {
				newConfig.Verbose = tVerbose
			}
		case "idOfMaster":
			tID, err := strconv.Atoi(argVal)
			if err != nil {
				fmt.Println(argVal, "is not a valid node id.")
				errorDetected = true
				break mainConfigLoop
			} else {
				newConfig.IdOfMaster = tID
			}
		case "configName":
			newConfig.ConfigName = argVal
		case "nodeType":
			switch strings.ToLower(argVal) {
			case "c", "client":
				newConfig.NodeType = "client"
			case "s", "server":
				newConfig.NodeType = "server"
			default:
				fmt.Println(argVal, "is not a valid node type.")
				errorDetected = true
				break mainConfigLoop
			}
		default:
			fmt.Println(argName, "is not a valid parameter.")
			errorDetected = true
			break mainConfigLoop
		}
	}

	if errorDetected {
		if !silent {
			fmt.Println("Configuration has not been modified.")
		}
		return false
	} else {
		*currentConfig = &newConfig
		if !silent {
			fmt.Println("New configuration accepted.")
		}
		return true
	}
}

func guidedConfig(scanner *bufio.Scanner, currentConfig **configBody) {
	var userInput string
	var args []string
	fmt.Println("Enter a name for this new configuration:")
	fmt.Println("White spaces are not supported.")
	for scanner.Scan() {
		fmt.Print("> ")
		userInput = "configName=" + scanner.Text()
		args = []string{userInput}
		if manualConfig(currentConfig, args, true) {
			break
		}
	}
	fmt.Println("What is the node type? (client/server)")
	for scanner.Scan() {
		fmt.Print("> ")
		userInput = "nodeType=" + scanner.Text()
		args = []string{userInput}
		if manualConfig(currentConfig, args, true) {
			break
		}
	}
	fmt.Println("Enter the node ID:")
	for scanner.Scan() {
		fmt.Print("> ")
		userInput = "nodeID=" + scanner.Text()
		args = []string{userInput}
		if manualConfig(currentConfig, args, true) {
			break
		}
	}
	fmt.Println("Enter the Node IPv4 address:")
	for scanner.Scan() {
		fmt.Print("> ")
		userInput = "nodeIP=" + scanner.Text()
		args = []string{userInput}
		if manualConfig(currentConfig, args, true) {
			break
		}
	}
	fmt.Println("Enter the Node Port number:")
	for scanner.Scan() {
		fmt.Print("> ")
		userInput = "nodePort=" + scanner.Text()
		args = []string{userInput}
		if manualConfig(currentConfig, args, true) {
			break
		}
	}
	tC := *currentConfig
	if tC.NodeType == "server" {
		fmt.Println("Enter the verbosity level (1:off/2:on):")
		for scanner.Scan() {
			fmt.Print("> ")
			userInput = "verbose=" + scanner.Text()
			args = []string{userInput}
			if manualConfig(currentConfig, args, true) {
				break
			}
		}
		fmt.Println("What is the ID of the master node?")
		for scanner.Scan() {
			fmt.Print("> ")
			userInput = "idOfMaster=" + scanner.Text()
			args = []string{userInput}
			if manualConfig(currentConfig, args, true) {
				break
			}
		}
	}
	fmt.Println("Guided configuration complete.")
}

func startNode(currentConfig *configBody) {
	if currentConfig.NodeType == "" {
		fmt.Println("Node type not specified!")
		return
	} else if currentConfig.NodeType == "server" {
		n := NodeComm.CreateNode(
			currentConfig.NodeID,
			currentConfig.IdOfMaster,
			currentConfig.NodeIP,
			currentConfig.NodePort,
			currentConfig.Verbose,
		)
		n.StartNode()
	} else if currentConfig.NodeType == "client" {
		c := Client.CreateClient(
			currentConfig.NodeID,
			currentConfig.NodeIP,
			currentConfig.NodePort,
		)
		c.StartClient()
	} else {
		fmt.Println("An unexpected error occured.")
	}
}

func saveConfig(currentConfig *configBody) {
	if currentConfig.ConfigName == "" {
		fmt.Println("Please provide the current configuration a name before attempting to save it.")
		return
	}
	_, err := os.Stat(configFolderPath)
	if err != nil || os.IsNotExist(err) {
		fmt.Println("Unable to access configuration folder. ", err)
		return
	}
	filePath := configFolderPath + currentConfig.ConfigName + ".json"
	_, err = os.Stat(filePath)
	if !(err != nil || os.IsNotExist(err)) {
		fmt.Println("A configuration by the name of", currentConfig.ConfigName, "already exists.")
		fmt.Println("Please either rename the current configuration or delete the existing saved configuration first.")
		return
	}
	fileToSave, err := json.MarshalIndent(*currentConfig, "", "	")
	if err != nil {
		fmt.Println("Error marshalling current configuration into a json for saving to file.", err)
		return
	}
	err = ioutil.WriteFile(filePath, fileToSave, 0644)
	if err != nil {
		fmt.Println("Error saving configuration to file.", err)
		return
	}
	fmt.Println("Configuration", currentConfig.ConfigName, "successfully saved.")
}

func loadConfig(currentConfig **configBody, args []string) {
	if len(args) < 1 {
		fmt.Println("No file specified.")
		return
	}
	filePath := configFolderPath + args[0] + ".json"
	_, err := os.Stat(filePath)
	if err != nil || os.IsNotExist(err) {
		fmt.Println("Unable to load", args[0], "as it does not exist.")
		return
	}
	loadedFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error loading", args[0], "from file.")
		return
	}
	var loadedConfig configBody
	json.Unmarshal(loadedFile, &loadedConfig)
	*currentConfig = &loadedConfig
	fmt.Println(args[0], "configuration successfully loaded.")
}

func listConfigs() {
	files, err := ioutil.ReadDir(configFolderPath)
	if err != nil {
		fmt.Println("Unable to find or access configuration folder. ", err)
		return
	}
	if len(files) == 0 {
		fmt.Println("Configuration folder is empty.")
		return
	}
	var configsFound []string
	for _, file := range files {
		if !file.IsDir() {
			tokensiedFileName := strings.Split(file.Name(), ".")
			if len(tokensiedFileName) != 2 {
				continue
			}
			if tokensiedFileName[1] != "json" {
				continue
			}
			configsFound = append(configsFound, tokensiedFileName[0])
		}
	}
	if len(configsFound) == 0 {
		fmt.Println("Configuration folder does not contain any valid configuration files.")
	} else {
		for _, configFound := range configsFound {
			fmt.Println(configFound)
		}
	}
}

func deleteConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("No configuration specified.")
	}
	filePath := configFolderPath + args[0] + ".json"
	_, err := os.Stat(filePath)
	if err != nil || os.IsNotExist(err) {
		fmt.Println("Unable to delete", args[0], "as it does not exist.")
		return
	}
	err = os.Remove(filePath)
	if err != nil {
		fmt.Println("Error deleting", args[0], ". ", err)
	}
	fmt.Println("Configuration", args[0], "deleted.")
}

func printConfig(currentConfig *configBody) {
	fmt.Println("Config Name:", currentConfig.ConfigName)
	fmt.Println("Node Type:", currentConfig.NodeType)
	fmt.Println("Node ID:", currentConfig.NodeID)
	fmt.Println("IPv4:", currentConfig.NodeIP)
	fmt.Println("Port:", currentConfig.NodePort)
	fmt.Println("Verbosity:", currentConfig.Verbose)
	if currentConfig.NodeType == "server" {
		fmt.Println("ID of Master:", currentConfig.IdOfMaster)
	}
	fmt.Println("Raw:", *currentConfig)
}

func printHelp(args []string) {
	if len(args) == 0 {
		fmt.Println("Available Commands:")
		fmt.Println("'help': \t\t\tPrints this help message or information about a given command.")
		fmt.Println("'config': \t\t\tUsed to set various parameters.")
		fmt.Println("'guided-config': \t\tStarts guided configuration.")
		fmt.Println("'start': \t\t\tStarts the node.")
		fmt.Println("'saveConfig': \t\t\tSaves the current configuration to file.")
		fmt.Println("'listConfigs': \t\t\tLists all known configurations.")
		fmt.Println("'loadConfig': \t\t\tLoads a specified configuration from file.")
		fmt.Println("'showConfig': \t\t\tPrints the current configuration.")
		fmt.Println("'printConfig': \t\t\tAn alias for 'showConfig'.")
		fmt.Println("'startConfig': \t\t\tLoads and starts a specified configuration.")
		fmt.Println("'updateDNS [IP] [PORT]': \tAllows you to update the DNS table'.")
		fmt.Println("'exit': \t\t\tExits the program.")
		return
	}
	switch args[0] {
	case "help":
		fmt.Println("'help' prints the standard help message or information about a given command.")
		fmt.Println("\tSyntax: help <command>")
	case "config":
		fmt.Println("'config' allows configuration parameters to be modified.")
		fmt.Println("\tConfiguration parameters should be specified as parameter_name=parameter_value")
		fmt.Println("\tSyntax: config parameter_name=parameter_value config parameter_name=parameter_value ...")
		fmt.Println("\tAllowed parameters:")
		fmt.Println("\t\tnodeID: Any integer value in the range [0, inf)")
		fmt.Println("\t\tnodeIP: Any valid IPv4 address.")
		fmt.Println("\t\tnodePort: Any valid port number.")
		fmt.Println("\t\tverbose: 1 (off) and 2 (on).")
		fmt.Println("\t\tidOfMaster: Node ID of master server.")
		fmt.Println("\t\tconfigName: A name for this configuration. Whitespaces are not supported.")
		fmt.Println("\t\tnodeType: 's' or 'server' for Server and 'c' or 'client' for client.")
	case "saveConfig":
		fmt.Println("'saveConfig' saves the current configuration to a .json file on disk.")
	case "listConfigs":
		fmt.Println("'listConfigs' lists all configurations in the configuration folder.")
	case "loadConfig":
		fmt.Println("'loadConfig' attempts to load the specified configuration from the configuration folder.")
		fmt.Println("\tSyntax: loadConfig <config_name>")
	default:
		fmt.Println(args[0], "either has no specific help text or it is an unknown command.")
	}
}

const configFolderPath = "config/"

type configBody struct {
	//Common fields
	NodeID   int
	NodeIP   string
	NodePort string
	Verbose  int

	//Server only fields
	IdOfMaster int

	//Config stuff
	ConfigName string
	NodeType   string
}
