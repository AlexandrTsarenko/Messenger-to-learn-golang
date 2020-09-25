package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"crypto/md5"

	"golang.org/x/crypto/ssh/terminal"
)

// debug variable
const debug = false //true

// NickName (after login)
var userNickName string

//
// Main
//
func main() {

	serverAddress := "localhost:1111"
	if len(os.Args) > 1 {
		serverAddress = os.Args[1]
	}

	fmt.Println("Connecting to '" + serverAddress + "' ...")

	// Connect
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Start reading routine
	go readRoutine(conn)

	// Print greeting
	fmt.Println(txtGREETING)

	//
	// Main client's loop
	//
	for {
		// Read user command from stdin
		fmt.Print(userNickName + "# ")
		command := readLine()
		if debug {
			log.Println("command:" + command)
		}

		// Process user command
		switch command {

		case cmdEXIT:
			return

		case cmdHELP:
			fmt.Print(txtHELP)

		case cmdREGISTER:
			handleRegister(conn)

		case cmdLOGIN:
			handleLogin(conn)

		case cmdLOGOUT:
			handleLogout(conn)

		case cmdLIST:
			handleList(conn)

		case cmdSEND:
			handleSend(conn)

		case cmdPASSWORD:
			handlePassword(conn)

		default:
			fmt.Println("Invalid command. (Use 'help' command).")
		}
	}
}

// handleRegister
func handleRegister(conn net.Conn) {
	//
	// obtain nickname
	//
	fmt.Print(" Enter your nickname:")
	nickName := readLine()
	if debug {
		log.Println("nickName: '" + nickName + "'")
	}

	// check unique nickname
	responseStr, err := sendRequest(conn, "CheckUniqueNickName", nickName, "")

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	//
	// obtain password
	//
	fmt.Print(" Enter password: ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("")

	// get md5 of password
	md5Hex := fmt.Sprintf("md5:%x", md5.Sum(bytePassword))
	if debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err = sendRequest(conn, "RegisterUser", nickName, md5Hex)

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}
}

// handleLogin
func handleLogin(conn net.Conn) {

	if userNickName != "" {
		fmt.Println("You are already logged")
		return
	}

	//
	// obtain nickname
	//
	fmt.Print(" Enter your nickname:")
	nickName := readLine()
	if debug {
		log.Println("nickName: '" + nickName + "'")
	}

	//
	// obtain password
	//
	fmt.Print(" Enter password: ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("")

	// get md5 of password
	md5Hex := fmt.Sprintf("md5:%x", md5.Sum(bytePassword))
	if debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err := sendRequest(conn, "Login", nickName, md5Hex)

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	userNickName = nickName
}

// handleLogout
func handleLogout(conn net.Conn) {

	// check authorization
	if userNickName == "" {
		fmt.Println("You are not logged.")
		return
	}

	// send request to server

	responseStr, err := sendRequest(conn, "Logout", "", "")

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	userNickName = ""
}

// handleList
func handleList(conn net.Conn) {

	// send request to server

	responseStr, err := sendRequest(conn, "GetOnlineUserList", "", "")

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	fmt.Println(responseStr)
}

// handleSend
func handleSend(conn net.Conn) {

	// check authorization
	if userNickName == "" {
		fmt.Println("You are not logged.")
		return
	}

	// get recipient name
	fmt.Print("to: ")
	recipientNickName := readLine()

	// get message text
	fmt.Print("enter message text: ")
	msgText := readLine()

	// send request to server

	responseStr, err := sendRequest(conn, "MessageTo", recipientNickName, msgText)

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}
}

// handlePassword
func handlePassword(conn net.Conn) {

	// check authorization
	if userNickName == "" {
		fmt.Println("You are not logged.")
		return
	}

	//
	// obtain new password
	//
	fmt.Print("Enter new password: ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
		return
	}

	// get md5 of password
	md5Hex := fmt.Sprintf("md5:%x", md5.Sum(bytePassword))
	if debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err := sendRequest(conn, "ChangePassword", md5Hex, "")

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}
}

// sendRequest
func sendRequest(conn net.Conn, command, data1, data2 string) (string, error) {

	// request structure data
	type Request struct {
		Command string
		Data1   string
		Data2   string
	}
	var requestData = Request{command, data1, data2}

	// encode to json
	requestBytes, err := json.Marshal(requestData)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	// send to server
	fmt.Fprintln(conn, string(requestBytes))
	if debug {
		log.Println("requestJson: " + string(requestBytes))
	}

	// wait response
	responseStr := <-responseChannel
	return responseStr, nil
}

// readLine from stdin
func readLine() string {
	rd := bufio.NewReader(os.Stdin)
	line, _ := rd.ReadString('\n')
	line = trimSuffix(line, "\n")
	return line
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func removeLastChar(s string) string {
	sz := len(s)
	if sz > 0 {
		s = s[:sz-1]
	}
	return s
}

// readingChannel
var responseChannel chan string = make(chan string)

//todo? var messagesChannel chan string = make(chan string)

//
// readRoutine
//
func readRoutine(conn net.Conn) {

	reader := bufio.NewReader(conn)

	// read loop
	for {
		//
		// read response
		//
		responseStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(10000000) // 0.01 sec
				continue
			}
			log.Println("read responnse err: " + err.Error())
			return
		}
		if debug {
			log.Print("   responseStr: " + responseStr)
		}

		// skip empty response (todo?)
		// if responseStr == "\n" {
		// 	continue
		// }

		//
		// receive message from other user
		//
		if responseStr == "message\n" {

			// get sender name
			from, err := reader.ReadString('\n')
			if debug {
				log.Print("   messageFrom: " + from)
			}
			if err != nil {
				log.Println("read message from err: " + err.Error())
				continue
			}
			from = removeLastChar(from)

			// get message text
			messageText, err := reader.ReadString('\n')
			if debug {
				log.Print("   messageText: " + messageText)
			}
			if err != nil {
				log.Println("read message text err: " + err.Error())
				continue
			}

			// print message to stdout
			fmt.Println("\n\nMessage from '" + from + "':\n" + messageText)

			// print new line
			fmt.Print(userNickName + "#")

			//todo? messagesChannel <- "Message from '" + from + "':\n" + messageText
			continue
		}

		//
		// receive response
		//
		if debug {
			log.Println("response: '" + removeLastChar(responseStr) + "'")
		}
		responseChannel <- removeLastChar(responseStr)
	}
}

// Commands
const (
	cmdEXIT     = "exit"
	cmdHELP     = "help"
	cmdREGISTER = "register"
	cmdLOGIN    = "login"
	cmdLOGOUT   = "logout"
	cmdLIST     = "list"
	cmdSEND     = "send"
	cmdPASSWORD = "password"
)

// Text constants
const txtGREETING = "Enter 'help' command to see a list of available commands\n"
const txtHELP = "" +
	"  '" + cmdREGISTER + "' - register new user\n" +
	"  '" + cmdLOGIN + "' - login existing user\n" +
	"  '" + cmdLOGOUT + "' - logout current user\n" +
	"  '" + cmdLIST + "' - get a list of online users\n" +
	"  '" + cmdSEND + "' - send a message to some user\n" +
	"  '" + cmdPASSWORD + "' - change password\n" +
	"  '" + cmdEXIT + "' - quit from this messager\n" +
	"  '" + cmdHELP + "' - display this help text\n"
