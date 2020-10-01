package client

import (
	"GitHub/Messenger-to-learn-golang/protocol"
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
const Debug = true //false //true

//
// Main
//
// func main() {

// 	serverAddress := "localhost:1111"

// 	if len(os.Args) > 1 {
// 		serverAddress = os.Args[1]
// 	}

// 	client := Client{}
// 	client.Run(serverAddress)
// 	return
// }

//
// Client - TCP client
//
type Client struct {

	// nickname (after login)
	userNickName string

	// connection to server
	conn net.Conn

	// feedback from readLoop go-routine
	responseChannel chan string
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
	go cl.readRoutine()

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

// handleExit
func (cl *Client) handleExit() {
	os.Exit(0)
}

// handleHelp
func (cl *Client) handleHelp() {
	fmt.Print(txtHELP)
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
	responseStr := cl.sendRequest(protocol.ScmdCheckUniqueNickName, nickName, "")

	if responseStr != "ok" {
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

	cl.sendRequest(protocol.ScmdRegisterUser, nickName, md5Hex)
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

	responseStr := cl.sendRequest(protocol.ScmdLogin, nickName, md5Hex)

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

	responseStr := cl.sendRequest(protocol.ScmdLogout, "", "")

	if responseStr != "ok" {
		return
	}

	cl.userNickName = ""
}

// handleList
func (cl *Client) handleList() {

	// send request to server

	responseStr := cl.sendRequest(protocol.ScmdGetOnlineUserList, "", "")

	if responseStr != "ok" {
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

	cl.sendRequest(protocol.ScmdMessageTo, recipientNickName, msgText)
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

	responseStr := cl.sendRequest(protocol.ScmdChangePassword, md5Hex, "")

	if responseStr != "ok" {
		fmt.Println(responseStr)
		return
	}
}

// sendRequest
func (cl *Client) sendRequest(command protocol.CommandToServer, data1, data2 string) string {

	// make requests json string
	requestData := protocol.Request{Command: command, Data1: data1, Data2: data2}
	requestStr := requestData.Encode()

	// send to server
	fmt.Fprintln(cl.conn, requestStr)
	if Debug {
		log.Println("requestJson: " + requestStr)
	}

	// wait response
	responseStr := <-cl.responseChannel

	if responseStr != "ok" {
		fmt.Println(responseStr)
	}

	return responseStr
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

//
// readRoutine
//
func (cl *Client) readRoutine() {

	reader := bufio.NewReader(cl.conn)

	// read loop
	for {
		//
		// read response or message from another user
		//
		responseStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(10000000) // 0.01 sec
				continue
			}
			log.Println("read err: " + err.Error())
			return
		}

		if Debug {
			log.Print("   responseStr: " + responseStr)
		}

		msg := protocol.MessageFromServer{}
		if err := msg.Decode(responseStr); err != nil {
			fmt.Println("Invalid message from server: " + responseStr)
			continue
		}

		switch msg.Type {

		case protocol.Reply:
			(cl.responseChannel) <- msg.ServerReply()

		case protocol.MessageFrom:
			// print message to stdout
			fmt.Println("\n\nMessage from '" + msg.SenderNickname() + "':\n" + msg.MessageText())

			// print new line
			fmt.Print(cl.userNickName + "#")
		}
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
