package pkg_test

import (
	"testing"
	"time"

	"git.internal.com/wingspan/pkg"
	"github.com/gorilla/websocket"
)

func TestServer(t *testing.T) {
	server := pkg.NewServer()
	if err := server.Register("Queue", pkg.NewQueue(2)); err != nil {
		t.Errorf("Could not register service: %v", err)
	}

	go server.Listen("0.0.0.0:8080")
	time.Sleep(time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080", nil)
	if err != nil {
		t.Fatalf("Could not connect to server: %v", err)
	}

	if err := conn.WriteJSON(pkg.Message{
		Method: "Queue.Add",
	}); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var response pkg.Response
	if err := conn.ReadJSON(&response); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
