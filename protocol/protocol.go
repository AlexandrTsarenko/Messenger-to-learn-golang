package protocol

import (
	"encoding/json"
	"log"
)

// MessageFromServer -
type MessageFromServer struct {
	Type  MessageType
	Data1 string
	Data2 string
}

// Encode - encodes MessageFromServer data structure to JSON string
func (r *MessageFromServer) Encode() string {

	// encode to json
	bytes, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	return string(bytes)
}

// Decode - decodes JSON string into MessageFromServer data structure
func (r *MessageFromServer) Decode(jsonStr string) error {
	json.Unmarshal([]byte(jsonStr), r)
	return nil
}

// MessageType - type of message from server
type MessageType string

const (
	// MtReply -
	MtReply MessageType = "Reply"

	// MtMessageFrom -
	MtMessageFrom MessageType = "MessageFrom"
)

//
// Request - data structure for the client to send a request to the server
//
type Request struct {
	Command CommandToServer
	Data1   string
	Data2   string
}

// Encode - encodes Request data structure to JSON string
func (r *Request) Encode() string {

	// encode to json
	bytes, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	return string(bytes)
}

// Decode - decodes JSON string into Request data structure
func (r *Request) Decode(jsonStr string) error {
	json.Unmarshal([]byte(jsonStr), r)
	return nil
}

// CommandToServer - command to server
type CommandToServer string

const (
	// ScmdCheckUniqueNickName - request to server
	ScmdCheckUniqueNickName CommandToServer = "CheckUniqueNickName"

	// ScmdRegisterUser - request to server
	ScmdRegisterUser CommandToServer = "RegisterUser"

	// ScmdLogin - request to server
	ScmdLogin CommandToServer = "Login"

	// ScmdLogout - request to server
	ScmdLogout CommandToServer = "Logout"

	// ScmdChangePassword - request to server
	ScmdChangePassword CommandToServer = "ChangePassword"

	// ScmdGetOnlineUserList - request to server
	ScmdGetOnlineUserList CommandToServer = "GetOnlineUserList"

	// ScmdMessageTo - request to server
	ScmdMessageTo CommandToServer = "MessageTo"

	// ScmdClear - request to server (for testing)
	ScmdClear CommandToServer = "Clear"
)
