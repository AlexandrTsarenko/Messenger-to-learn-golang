package main

import (
	"GitHub/Messenger-to-learn-golang/protocol"
	"net"
	"testing"
)

func TestRegister(t *testing.T) {
	var cl = Client{}

	var err error
	cl.conn, err = net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No Sever Connection: ", err)
		return
	}

	cl.responseChannel = make(chan string)
	go cl.readRoutine()

	// Clear
	r := cl.sendRequest(protocol.ScmdClear, "", "")

	// RegisterUser 'a'
	r = cl.sendRequest(protocol.ScmdRegisterUser, "a", "md5")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login/Logout
	r = cl.sendRequest(protocol.ScmdLogin, "a", "md5")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// Logout
	r = cl.sendRequest(protocol.ScmdLogout, "a", "")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	cl.conn.Close()
	cl.conn, err = net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No Sever Connection: ", err)
		return
	}
	go cl.readRoutine()

	// Invalid password
	r = cl.sendRequest(protocol.ScmdLogin, "a", "pass")
	if r != "Invalid password" {
		t.Error("Response error: ", r)
		return
	}

	// Login/+Login
	r = cl.sendRequest(protocol.ScmdLogin, "a", "md5")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// +Login
	r = cl.sendRequest(protocol.ScmdLogin, "a", "md5")
	if r != "User 'a' is already online" {
		t.Error("Response error: ", r)
		return
	}

	// Again RegisterUser 'a'
	r = cl.sendRequest(protocol.ScmdRegisterUser, "a", "md5")
	if r != "User 'a' already exists" {
		t.Error("Response error: ", r)
		return
	}

	cl2 := Client{}
	// conn2
	cl2.conn, err = net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No 2-th Sever Connection: ", err)
		return
	}
	cl2.responseChannel = make(chan string)
	go cl2.readRoutine()

	// RegisterUser 'b'
	r = cl2.sendRequest(protocol.ScmdRegisterUser, "b", "md5")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login (b)
	r = cl2.sendRequest(protocol.ScmdLogin, "b", "md5")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Get List (b)
	r = cl2.sendRequest(protocol.ScmdGetOnlineUserList, "", "")
	if r != "online users: a,b" {
		t.Error("Response error: ", r)
		return
	}

	// Send Message
	// go func() {
	// 	responseChannel <- "ok"
	// }()
	r = cl.sendRequest(protocol.ScmdMessageTo, "b", "Msg")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// msg := <-messagesChannel
	// if msg != "msg" {
	// 	t.Error("Response error: ", r)
	// 	return
	// }

	// Clear
	r = cl.sendRequest(protocol.ScmdClear, "", "")
}
