package pkg_test

import (
	"testing"
	"time"

	"git.internal.com/wingspan/pkg"
	"github.com/gorilla/websocket"
)

type FakeMatchmaker struct{}

func (f *FakeMatchmaker) CreateMatch(socket *pkg.Sockt, players []pkg.Socket) (*pkg.Message, error) {
	response := pkg.Response{Type: "fake_match_created"}
	for _, player := range players {
		if n, err := player.Send(response); n > 0 && err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func TestServer(t *testing.T) {
	t.Run("send message", func(t *testing.T) {
		server := pkg.NewServer()
		defer server.Close()

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
	})

	t.Run("reply", func(t *testing.T) {
		server := pkg.NewServer()
		defer server.Close()

		server.Register("Queue", pkg.NewQueue(2))
		server.Register("Matchmaker", new(FakeMatchmaker))

		go server.Listen("0.0.0.0:8080")
		time.Sleep(time.Millisecond)

		p1, _, _ := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080", nil)
		p2, _, _ := websocket.DefaultDialer.Dial("ws://0.0.0.0:8080", nil)

		p1.WriteJSON(pkg.Message{Method: "Queue.Add"})
		p1.ReadJSON(nil)

		p2.WriteJSON(pkg.Message{Method: "Queue.Add"})
		p2.ReadJSON(nil)

		var response pkg.Response
		if err := p1.ReadJSON(&response); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if response.Type != "fake_match_created" {
			t.Errorf("Expected type %v, got %v", "fake_match_created", response.Type)
		}
	})
}
