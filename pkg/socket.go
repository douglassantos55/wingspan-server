package pkg

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

type Socket interface {
	io.ReadWriteCloser
	// Helper to send responses instead of handling io
	Send(response Response) (int, error)
}

type Sockt struct {
	conn     *websocket.Conn
	Incoming chan Message
	Outgoing chan Response
}

func NewSocket(conn *websocket.Conn) *Sockt {
	socket := &Sockt{
		conn:     conn,
		Incoming: make(chan Message),
		Outgoing: make(chan Response),
	}

	go func() {
		defer socket.Close()
		for {
			data, err := io.ReadAll(socket)
			if err != nil {
				break
			}
			var message Message
			if err := json.Unmarshal(data, &message); err != nil {
				log.Printf("Could not decode message: %v", message)
				continue
			}
			socket.Incoming <- message
		}
	}()

	return socket
}

func (s *Sockt) Send(response Response) (int, error) {
	data, err := json.Marshal(response)
	if err != nil {
		return 0, err
	}
	return s.Write(data)
}

func (s *Sockt) Close() error {
	close(s.Incoming)
	close(s.Outgoing)
	return s.conn.Close()
}

func (s *Sockt) Write(data []byte) (int, error) {
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, err
	}
	if err := s.conn.WriteJSON(response); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (s *Sockt) Read(data []byte) (int, error) {
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

type TestSocket struct {
	buf *RingBuffer[*bytes.Buffer]
}

func NewTestSocket() *TestSocket {
	return &TestSocket{
		buf: NewRingBuffer[*bytes.Buffer](10),
	}
}

func (t *TestSocket) Close() error {
	return nil
}

func (t *TestSocket) Read(p []byte) (int, error) {
	buffer := t.buf.Last()
	if buffer == nil {
		return 0, io.EOF
	}
	return buffer.Read(p)
}

func (t *TestSocket) Write(p []byte) (int, error) {
	buffer := new(bytes.Buffer)
	n, err := buffer.Write(p)
	t.buf.Push(buffer)
	return n, err
}

func (t *TestSocket) Send(response Response) (int, error) {
	data, err := json.Marshal(response)
	if err != nil {
		return 0, err
	}
	return t.Write(data)
}

// Helper to receive responses instead of handling io
func (t *TestSocket) GetResponse() (*Response, error) {
	defer t.buf.Pop()
	data, err := io.ReadAll(t)
	if err != nil {
		return nil, err
	}
	var response *Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return response, nil
}
