package protocol

import (
	"encoding/json"
	"log"
)

// MessageFromServer - message from server to client
type MessageFromServer struct {
	Type  MessageType
	Data1 string
	Data2 string
}

// ServerReply -
func (m *MessageFromServer) ServerReply() string {
	return m.Data1
}

// SenderNickname -
func (m *MessageFromServer) SenderNickname() string {
	return m.Data1
}

// MessageText -
func (m *MessageFromServer) MessageText() string {
	return m.Data2
}

// Encode - encodes 'MessageFromServer data structure' to 'JSON string'
func (m *MessageFromServer) Encode() []byte {

	// encode to json
	bytes, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
		return []byte{}
	}

	return bytes
}

// Decode - decodes 'JSON string' into 'MessageFromServer data structure'
func (m *MessageFromServer) Decode(jsonStr string) error {
	json.Unmarshal([]byte(jsonStr), m)
	return nil
}

// MessageType - type of message from server
type MessageType string

const (
	// Reply -
	Reply MessageType = "Reply"

	// MessageFrom -
	MessageFrom MessageType = "MessageFrom"
)
