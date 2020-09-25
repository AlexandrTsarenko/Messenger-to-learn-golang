package main

import (
	"net"
	"testing"
)

func TestRegister(t *testing.T) {

	conn, err := net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No Sever Connection: ", err)
		return
	}
	go readRoutine(conn)

	// Clear
	r, err := sendRequest(conn, "Clear", "", "")
	if err != nil {
		t.Error("Error: ", err)
		return
	}

	// RegisterUser 'a'
	r, err = sendRequest(conn, "RegisterUser", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login/Logout
	r, err = sendRequest(conn, "Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// Logout
	r, err = sendRequest(conn, "Logout", "a", "")
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	conn.Close()
	conn, err = net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No Sever Connection: ", err)
		return
	}
	go readRoutine(conn)

	// Invalid password
	r, err = sendRequest(conn, "Login", "a", "pass")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "Invalid password" {
		t.Error("Response error: ", r)
		return
	}

	// Login/+Login
	r, err = sendRequest(conn, "Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}
	// +Login
	r, err = sendRequest(conn, "Login", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "User 'a' is already online" {
		t.Error("Response error: ", r)
		return
	}

	// Again RegisterUser 'a'
	r, err = sendRequest(conn, "RegisterUser", "a", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "User 'a' already exists" {
		t.Error("Response error: ", r)
		return
	}

	// conn2
	conn2, err := net.Dial("tcp", "localhost:1111")
	if err != nil {
		t.Error("No 2-th Sever Connection: ", err)
		return
	}
	go readRoutine(conn2)

	// RegisterUser 'b'
	r, err = sendRequest(conn2, "RegisterUser", "b", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Login (b)
	r, err = sendRequest(conn2, "Login", "b", "md5")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
	if r != "ok" {
		t.Error("Response error: ", r)
		return
	}

	// Get List (b)
	r, err = sendRequest(conn2, "GetOnlineUserList", "", "")
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
	r, err = sendRequest(conn, "MessageTo", "b", "Msg")
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
	r, err = sendRequest(conn, "Clear", "", "")
	if err != nil {
		t.Error("Error: ", err)
		return
	}
}
