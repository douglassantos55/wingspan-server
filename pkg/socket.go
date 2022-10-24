package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

type Socket struct {
	conn     *websocket.Conn
	Incoming chan Message
	Outgoing chan Response
}

func NewTestSocket() io.ReadWriter {
	return &bytes.Buffer{}
}

func NewSocket(conn *websocket.Conn) *Socket {
	socket := &Socket{
		conn:     conn,
		Incoming: make(chan Message),
		Outgoing: make(chan Response),
	}

	go func() {
		for {
			data, err := io.ReadAll(socket)
			if err != nil {
				continue
			}
			var message Message
			if err := json.Unmarshal(data, &message); err != nil {
				log.Printf("Could not decode message: %v", message)
				continue
			}
			socket.Incoming <- message
		}
	}()

	go func() {
		for {
			response := <-socket.Outgoing
			fmt.Printf("response: %v\n", response)
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("Could not encode response: %v", response)
				continue
			}

			if _, err := socket.Write(data); err != nil {
				log.Printf("Could not write response: %v", response)
				continue
			}
		}
	}()

	return socket
}

func (s *Socket) Write(data []byte) (int, error) {
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, err
	}
	if err := s.conn.WriteJSON(response); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (s *Socket) Read(data []byte) (int, error) {
	var message Message
	if err := s.conn.ReadJSON(&message); err != nil {
		return 0, err
	}

	encoding, err := json.Marshal(message)
	if err != nil {
		return 0, err
	}

	return copy(data, encoding), io.EOF
}
