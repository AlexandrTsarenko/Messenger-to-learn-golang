package server_test

import (
	"GitHub/Messenger-to-learn-golang/server"
	"log"
	"testing"
)

func TestServer(t *testing.T) {

	log.SetFlags( /*log.LstdFlags |*/ log.Lshortfile)

	server.Debug = true

	srv := server.NewServer(":1111")

	srv.Run()
}
