package main

import (
	"GitHub/Messenger-to-learn-golang/protocol"
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

// Debug - for Debugging
const Debug = true //false //true

// LocalDbInterface - Interface for local DB implementaion
type LocalDbInterface interface {
	// Init - Initiate Local Db
	Init() error

	// RLock - lock for reading
	RLock()

	// RUnlock - unlock for reading
	RUnlock()

	// FindUser - find user
	FindUser(name string) (*UserInfo, bool)

	// DoesUserExist - check that user exists
	DoesUserExist(name string) bool

	// AddUser - Add User
	AddUser(name, password string, conn net.Conn) error

	// Login - login
	Login(name, password string, conn net.Conn) error

	// Logout - logout
	Logout(name string)

	// ChangePassword -
	ChangePassword(name, newPassword string) error

	// GetOnlineUserList - Get Online User List
	GetOnlineUserList() []string

	// Clear - Clear Local Db (for testing)
	Clear()
}

// Local DB
var gLocalDb LocalDbInterface = &LocalDb{}

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

// Server - TCP message server
type Server struct {
	port    string
	localDb LocalDbInterface
}

// Run -
func (srv *Server) Run() {

	srv.localDb.Init()

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

// handleConnection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	//log.Printf("Serving %s\n", conn.RemoteAddr().String())

	// user name after login
	userName := ""

	for {

		// read client request
		requestStr, err := bufio.NewReader(conn).ReadString('\n')

		// connection lost?
		if err != nil {
			if Debug {
				log.Println("userName: '" + userName)
				log.Println(err)
			}
			if userName != "" {
				gLocalDb.Logout(userName)
				userName = ""
			}
			return
		}

		if Debug {
			log.Print("requestStr: " + requestStr + "\n")
		}

		// decode to request data
		var rqst protocol.Request
		rqst.Decode(requestStr)

		//
		// Process client request
		//
		switch rqst.Command {

		//  CheckUniqueNickName
		case protocol.ScmdCheckUniqueNickName:
			if gLocalDb.DoesUserExist(rqst.Data1) {
				sendReply(conn, "User '"+rqst.Data1+"' already exists")

			} else {
				sendReply(conn, "ok")
			}

		//  RegisterUser
		case protocol.ScmdRegisterUser:
			if err := gLocalDb.AddUser(rqst.Data1, rqst.Data2, conn); err != nil {
				sendReply(conn, err.Error())
			} else {
				sendReply(conn, "ok")
			}

		//  Login
		case protocol.ScmdLogin:
			if err := gLocalDb.Login(rqst.Data1, rqst.Data2, conn); err != nil {
				sendReply(conn, err.Error())
			} else {
				userName = rqst.Data1
				sendReply(conn, "ok")
			}

		//  Logout
		case protocol.ScmdLogout:
			gLocalDb.Logout(userName)
			userName = ""
			sendReply(conn, "ok")

		//  ChangePassword
		case protocol.ScmdChangePassword:
			if err := gLocalDb.ChangePassword(userName, rqst.Data1); err != nil {
				sendReply(conn, err.Error())
			} else {
				sendReply(conn, "ok")
			}

		//  GetOnlineUserList
		case protocol.ScmdGetOnlineUserList:
			userList := gLocalDb.GetOnlineUserList()
			if len(userList) > 0 {
				sendReply(conn, "online users: "+strings.Join(userList, ","))
			} else {
				sendReply(conn, "no online users")
			}

		//  MessageTo
		case protocol.ScmdMessageTo:

			// check login status
			if userName == "" {
				sendReply(conn, "You are not logged in")
				continue
			}

			gLocalDb.RLock()
			{
				// get recipient user info
				name := rqst.Data1
				userInfo, isFound := gLocalDb.FindUser(name)

				// check recipient connection
				if !isFound {
					sendReply(conn, "User '"+name+"' does not exist")
				} else if userInfo.conn == nil {
					sendReply(conn, "User '"+name+"' is offline")
				} else {

					// send message
					if err := sendMessage(userInfo.conn, userName, rqst.Data2); err != nil {
						sendReply(conn, err.Error())
					} else {
						sendReply(conn, "ok")
					}
				}
			}
			gLocalDb.RUnlock()

		//  Clear (for testing)
		case protocol.ScmdClear:
			gLocalDb.Clear()
			sendReply(conn, "ok")
		}
	} // end of for
}

// Forward message from one user to another
func sendReply(conn net.Conn, replyText string) error {
	msg := protocol.MessageFromServer{Type: protocol.Reply, Data1: replyText, Data2: ""}
	json := append(msg.Encode(), '\n')
	_, err := conn.Write(json)
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}

// Forward message from one user to another
func sendMessage(conn net.Conn, name, message string) error {
	msg := protocol.MessageFromServer{Type: protocol.MessageFrom, Data1: name, Data2: message}
	json := append(msg.Encode(), '\n')
	_, err := conn.Write(json)
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}
