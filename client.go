package main

import (
	"GitHub/Messenger-to-learn-golang/request"
	"bufio"
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

// Debug - for debugging
const Debug = false //true

//
// Main
//
func main() {

	serverAddress := "localhost:1111"

	if len(os.Args) > 1 {
		serverAddress = os.Args[1]
	}

	client := Client{}
	client.Run(serverAddress)
	return
}

//
// Client - TCP client
//
type Client struct {

	// nickname (after login)
	userNickName string

	// feedback from readLoop go-routine
	responseChannel chan string

	// connection to server
	conn net.Conn
}

//
// Run - run TCP client
//
func (cl *Client) Run(serverAddr string) {

	fmt.Println("Connecting to '" + serverAddr + "' ...")

	// Connect to server
	var err error
	cl.conn, err = net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cl.conn.Close()

	// Start response reading routine
	cl.responseChannel = make(chan string)
	go cl.readRoutine(cl.conn)

	// Print greeting
	fmt.Println(txtGREETING)

	var commandMap = map[string]func(){
		cmdEXIT:     cl.handleExit,
		cmdHELP:     cl.handleHelp,
		cmdREGISTER: cl.handleRegister,
		cmdLOGIN:    cl.handleLogin,
		cmdLOGOUT:   cl.handleLogout,
		cmdLIST:     cl.handleList,
		cmdMESSAGE:  cl.handleSendMessage,
		cmdPASSWORD: cl.handlePassword,
	}

	//
	// Command line read loop
	//
	for {
		// Read user command from stdin
		fmt.Print(cl.userNickName + "# ")
		command := readLine()
		if Debug {
			log.Println("command:" + command)
		}

		if handleFunc := commandMap[command]; handleFunc != nil {
			handleFunc()
		} else {
			fmt.Println("Invalid command. (Use 'help' command).")
		}
	}
}

// handleRegister
func (cl *Client) handleRegister() {
	//
	// obtain nickname
	//
	fmt.Print(" Enter your nickname:")
	nickName := readLine()
	if Debug {
		log.Println("nickName: '" + nickName + "'")
	}

	// check unique nickname
	responseStr, err := cl.sendRequest("CheckUniqueNickName", nickName, "")

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
	if Debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err = cl.sendRequest("RegisterUser", nickName, md5Hex)

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
func (cl *Client) handleLogin() {

	if cl.userNickName != "" {
		fmt.Println("You are already logged")
		return
	}

	//
	// obtain nickname
	//
	fmt.Print(" Enter your nickname:")
	nickName := readLine()
	if Debug {
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
	if Debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err := cl.sendRequest("Login", nickName, md5Hex)

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	cl.userNickName = nickName
}

// handleLogout
func (cl *Client) handleLogout() {

	// check authorization
	if cl.userNickName == "" {
		fmt.Println("You are not logged.")
		return
	}

	// send request to server

	responseStr, err := cl.sendRequest("Logout", "", "")

	if err != nil {
		log.Fatal(err)
		return
	}

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}

	cl.userNickName = ""
}

// handleList
func (cl *Client) handleList() {

	// send request to server

	responseStr, err := cl.sendRequest("GetOnlineUserList", "", "")

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

// handleSendMessage
func (cl *Client) handleSendMessage() {

	// check authorization
	if cl.userNickName == "" {
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

	responseStr, err := cl.sendRequest("MessageTo", recipientNickName, msgText)

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
func (cl *Client) handlePassword() {

	// check authorization
	if cl.userNickName == "" {
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
	if Debug {
		log.Print("md5:" + md5Hex)
	}

	// send request to server

	responseStr, err := cl.sendRequest("ChangePassword", md5Hex, "")

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
func (cl *Client) sendRequest(command, data1, data2 string) (string, error) {

	// make requests json string
	requestData := request.Request{Command: command, Data1: data1, Data2: data2}
	requestStr := requestData.Encode()

	// send to server
	fmt.Fprintln(cl.conn, requestStr)
	if Debug {
		log.Println("requestJson: " + requestStr)
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

// func removeLastChar(s string) string {
// 	sz := len(s)
// 	if sz > 0 {
// 		s = s[:sz-1]
// 	}
// 	return s
// }

// readingChannel
var responseChannel chan string = make(chan string)

//todo? var messagesChannel chan string = make(chan string)

//
// readRoutine
//
func (cl *Client) readRoutine(conn net.Conn) {

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
		if Debug {
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
			if Debug {
				log.Print("   messageFrom: " + from)
			}
			if err != nil {
				log.Println("read message from err: " + err.Error())
				continue
			}
			from = trimSuffix(from, "\n")

			// get message text
			messageText, err := reader.ReadString('\n')
			if Debug {
				log.Print("   messageText: " + messageText)
			}
			if err != nil {
				log.Println("read message text err: " + err.Error())
				continue
			}

			// print message to stdout
			fmt.Println("\n\nMessage from '" + from + "':\n" + messageText)

			// print new line
			fmt.Print(cl.userNickName + "#")

			//todo? messagesChannel <- "Message from '" + from + "':\n" + messageText
			continue
		}

		//
		// receive response
		//
		if Debug {
			log.Println("response: '" + trimSuffix(responseStr, "\n") + "'")
		}
		responseChannel <- trimSuffix(responseStr, "\n")
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
	cmdMESSAGE  = "send"
	cmdPASSWORD = "password"
)

func (cl *Client) handleExit() {
	os.Exit(0)
}

func (cl *Client) handleHelp() {
	fmt.Print(txtHELP)
}

// Text constants
const txtGREETING = "Enter 'help' command to see a list of available commands\n"
const txtHELP = "" +
	"  '" + cmdREGISTER + "' - register new user\n" +
	"  '" + cmdLOGIN + "' - login existing user\n" +
	"  '" + cmdLOGOUT + "' - logout current user\n" +
	"  '" + cmdLIST + "' - get a list of online users\n" +
	"  '" + cmdMESSAGE + "' - send a message to some user\n" +
	"  '" + cmdPASSWORD + "' - change password\n" +
	"  '" + cmdEXIT + "' - quit from this messager\n" +
	"  '" + cmdHELP + "' - display this help text\n"
