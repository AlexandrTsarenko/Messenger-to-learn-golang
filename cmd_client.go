package main

import (
	"GitHub/Messenger-to-learn-golang/client"
	"os"
)

func main() {

	serverAddress := "localhost:1111"

	if len(os.Args) > 1 {
		serverAddress = os.Args[1]
	}

	client := client.Client{}
	client.Run(serverAddress)
	return
}
