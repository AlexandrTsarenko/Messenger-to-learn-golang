package main

import (
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

	go cl.readRoutine(cl.conn)

	// Clear
	r, err := cl.sendRequest("Clear", "", "")
	if err != nil {
		t.Error("Error: ", err)
		return
	}

	// RegisterUser 'a'
	r, err = cl.sendRequest("RegisterUser", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login/Logout
	r, err = cl.sendRequest("Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// Logout
	r, err = cl.sendRequest("Logout", "a", "")
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
	go cl.readRoutine(cl.conn)

	// Invalid password
	r, err = cl.sendRequest("Login", "a", "pass")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "Invalid password" {
		t.Error("Response error: ", r)
		return
	}

	// Login/+Login
	r, err = cl.sendRequest("Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// +Login
	r, err = cl.sendRequest("Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "User 'a' is already online" {
		t.Error("Response error: ", r)
		return
	}

	// Again RegisterUser 'a'
	r, err = cl.sendRequest("RegisterUser", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
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
	go cl2.readRoutine(cl2.conn)

	// RegisterUser 'b'
	r, err = cl2.sendRequest("RegisterUser", "b", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login (b)
	r, err = cl2.sendRequest("Login", "b", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Get List (b)
	r, err = cl2.sendRequest("GetOnlineUserList", "", "")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "online users: a,b" {
		t.Error("Response error: ", r)
		return
	}

	// Send Message
	// go func() {
	// 	responseChannel <- "ok"
	// }()
	r, err = cl.sendRequest("MessageTo", "b", "Msg")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
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
	r, err = cl.sendRequest("Clear", "", "")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
}
