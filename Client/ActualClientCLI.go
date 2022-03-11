package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//Implement a CLI here that is built on top of the ActualClientAPI.go
func (C *Client) startCLI() {
	scanner := bufio.NewScanner(os.Stdin)
	var userInput string
	fmt.Print("> ")

Main:
	for scanner.Scan() {
		userInput = ""
		userInput = scanner.Text()
		tokenised := strings.Fields(userInput)

		if len(tokenised) == 0 {
			continue
		}

		switch tokenised[0] {
		case "exit":
			break Main
		case "write":
			C.ClientRequest("Write")
		case "read":
			C.ClientRequest("Read")
		case "help":
			printHelp(tokenised)
		default:
			fmt.Printf("Invalid Input: %s\n", userInput)
		}

		fmt.Print("> ")
	}
}

func printHelp(params []string) {
	if len(params) == 1 {
		fmt.Println("'exit':\t Exit program.")
		fmt.Println("'writeRequest':\t Client sends Write Request to Master.")
		fmt.Println("'readRequest':\t Client sends Read Request to Master.")
	}
}
