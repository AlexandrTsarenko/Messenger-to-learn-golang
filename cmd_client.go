package main

import (
	"os"

	"GitHub/Messenger-to-learn-golang/client"
)

// Debug - for debugging
const Debug = true //false //true

//
// Main
//
func main() {

	serverAddress := "localhost:1111"

	if len(os.Args) > 1 {
		serverAddress = os.Args[1]
	}

	client := client.Client{}
	client.Run(serverAddress)
	return
}
