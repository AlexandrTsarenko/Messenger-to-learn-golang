package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
)

// for debugging
const debug = false //true

// Local DB
var gLocalDb LocalDbStruct

//
// main()
//
func main() {

	log.SetFlags( /*log.LstdFlags |*/ log.Lshortfile)

	gLocalDb.Init()

	portNum := ":1111"
	if len(os.Args) > 1 {
		portNum = os.Args[1]
	}

	listener, err := net.Listen("tcp4", portNum)
	if err != nil {
		log.Println("Listen failed!")
		log.Fatal(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept failed!")
			log.Println(err)
			return
		}
		go handleConnection(conn)
	}
}

/// handleConnection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	//log.Printf("Serving %s\n", conn.RemoteAddr().String())

	// user name after login
	userName := ""

	for {

		// client request data
		type Request struct {
			Command string
			Data1   string
			Data2   string
		}
		var request Request

		// read client request
		requestStr, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if debug {
				log.Println("userName: '" + userName)
				log.Println(err)
			}
			if userName != "" {
				gLocalDb.Logout(userName)
				userName = ""
			}
			// connection lost?
			return
		}
		if debug {
			log.Print("requestStr: " + requestStr + "\n")
		}
		json.Unmarshal([]byte(requestStr), &request)

		//
		// Process client request
		//
		switch request.Command {

		//  CheckUniqueNickName
		case "CheckUniqueNickName":
			if gLocalDb.DoesUserExist(request.Data1) {
				conn.Write([]byte("User '" + request.Data1 + "' already exists\n"))
			} else {
				conn.Write([]byte("ok\n"))
			}

		//  RegisterUser
		case "RegisterUser":
			if err := gLocalDb.AddUser(request.Data1, request.Data2, conn); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("ok\n"))
			}

		//  Login
		case "Login":
			if err := gLocalDb.Login(request.Data1, request.Data2, conn); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				userName = request.Data1
				conn.Write([]byte("ok\n"))
			}

		//  Logout
		case "Logout":
			gLocalDb.Logout(userName)
			userName = ""
			conn.Write([]byte("ok\n"))

		//  ChangePassword
		case "ChangePassword":
			if err := gLocalDb.ChangePassword(userName, request.Data1); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("ok\n"))
			}

		//  GetOnlineUserList
		case "GetOnlineUserList":
			userList := gLocalDb.GetOnlineUserList()
			if len(userList) > 0 {
				conn.Write([]byte("online users: " + strings.Join(userList, ",") + "\n"))
			} else {
				conn.Write([]byte("no online users\n"))
			}

		//  MessageTo
		case "MessageTo":

			// check login status
			if userName == "" {
				conn.Write([]byte("You are not logged in\n"))
				continue
			}

			gLocalDb.RLock()
			{
				// get recipient user info
				name := request.Data1
				userInfo, isFound := gLocalDb.FindUser(name)

				// check recipient connection
				if !isFound {
					conn.Write([]byte("User '" + name + "' does not exist\n"))
				} else if userInfo.conn == nil {
					conn.Write([]byte("User '" + name + "' is offline\n"))
				} else {

					// send message
					if err := sendMessage(userInfo.conn, userName, request.Data2); err != nil {
						conn.Write([]byte(err.Error() + "\n"))
					} else {
						conn.Write([]byte("ok\n"))
					}
				}
			}
			gLocalDb.RUnlock()

		//  Clear (for testing)
		case "Clear":
			if debug {
				gLocalDb.Clear()
				conn.Write([]byte("ok\n"))
			}
		}
	} // end of for
}

// Forward message from one user to another
func sendMessage(conn net.Conn, name, message string) error {
	_, err := conn.Write([]byte("message\n" + name + "\n" + message + "\n"))
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}
