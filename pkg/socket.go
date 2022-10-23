package pkg

import (
	"bytes"
	"encoding/gob"
	"io"
	"io/ioutil"
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
			data, err := ioutil.ReadAll(socket)
			if err != nil {
				log.Printf("Could not read message: %v", err)
				continue
			}

			buf := bytes.NewBuffer(data)
			decoder := gob.NewDecoder(buf)

			var message Message
			if err := decoder.Decode(&message); err != nil {
				log.Printf("Could not decode message: %v", message)
				continue
			}
			socket.Incoming <- message
		}
	}()

	go func() {
		for {
			response := <-socket.Outgoing

			buf := bytes.Buffer{}
			encoder := gob.NewEncoder(&buf)

			if err := encoder.Encode(response); err != nil {
				log.Printf("Could not encode response: %v", response)
				continue
			}

			if _, err := socket.Write(buf.Bytes()); err != nil {
				log.Printf("Could not write response: %v", response)
				continue
			}
		}
	}()

	return socket
}

func (s *Socket) Write(data []byte) (int, error) {
	if err := s.conn.WriteJSON(data); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (s *Socket) Read(data []byte) (int, error) {
	if err := s.conn.ReadJSON(&data); err != nil {
		return 0, err
	}
	return len(data), nil
}
