package main

import (
	"GitHub/Messenger-to-learn-golang/request"
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

// for Debugging
const Debug = false //true

// LocalDbInterface - Interface for local DB implementaion
type LocalDbMethods interface {
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
var gLocalDb LocalDbMethods = &LocalDb{}

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
		var rqst request.Request
		rqst.Decode(requestStr)

		//
		// Process client request
		//
		switch rqst.Command {

		//  CheckUniqueNickName
		case "CheckUniqueNickName":
			if gLocalDb.DoesUserExist(rqst.Data1) {
				conn.Write([]byte("User '" + rqst.Data1 + "' already exists\n"))
			} else {
				conn.Write([]byte("ok\n"))
			}

		//  RegisterUser
		case "RegisterUser":
			if err := gLocalDb.AddUser(rqst.Data1, rqst.Data2, conn); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				conn.Write([]byte("ok\n"))
			}

		//  Login
		case "Login":
			if err := gLocalDb.Login(rqst.Data1, rqst.Data2, conn); err != nil {
				conn.Write([]byte(err.Error() + "\n"))
			} else {
				userName = rqst.Data1
				conn.Write([]byte("ok\n"))
			}

		//  Logout
		case "Logout":
			gLocalDb.Logout(userName)
			userName = ""
			conn.Write([]byte("ok\n"))

		//  ChangePassword
		case "ChangePassword":
			if err := gLocalDb.ChangePassword(userName, rqst.Data1); err != nil {
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
				name := rqst.Data1
				userInfo, isFound := gLocalDb.FindUser(name)

				// check recipient connection
				if !isFound {
					conn.Write([]byte("User '" + name + "' does not exist\n"))
				} else if userInfo.conn == nil {
					conn.Write([]byte("User '" + name + "' is offline\n"))
				} else {

					// send message
					if err := sendMessage(userInfo.conn, userName, rqst.Data2); err != nil {
						conn.Write([]byte(err.Error() + "\n"))
					} else {
						conn.Write([]byte("ok\n"))
					}
				}
			}
			gLocalDb.RUnlock()

		//  Clear (for testing)
		case "Clear":
			gLocalDb.Clear()
			conn.Write([]byte("ok\n"))
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
