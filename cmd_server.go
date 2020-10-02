package main

import (
	"GitHub/Messenger-to-learn-golang/server"
	"os"
)

func main() {

	portNum := ":1111"
	if len(os.Args) > 1 {
		portNum = os.Args[1]
	}

	srv := server.NewServer(portNum)

	srv.Run()
}
