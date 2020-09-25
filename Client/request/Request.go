package request

import (
	"encoding/json"
	"log"
	"net"
)

// Request - data structure for the client to send a request to the server
type Request struct {
	Command string
	Data1   string
	Data2   string
}

// Endode - encodes Request data structure to JSON string
func (r *Request) Endode(conn net.Conn) string {

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
