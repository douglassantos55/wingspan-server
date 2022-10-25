package pkg_test

import (
	"encoding/json"
	"io"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestMatchmaker(t *testing.T) {
	t.Run("accept match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch([]pkg.Socket{p1, p2})
		reply, err := matchmaker.Accept(p1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if reply != nil {
			t.Errorf("Expected no response, got %v", reply)
		}

		data, err := io.ReadAll(p1)
		if err != nil {
			t.Fatalf("Could not read data from socket: %v", err)
		}

		var response pkg.Response
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatalf("Could not parse response: %v", err)
		}
		if response.Type != pkg.WaitOtherPlayers {
			t.Errorf("Expected type %v, got %v", pkg.WaitOtherPlayers, response.Type)
		}
	})

	t.Run("start game", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch([]pkg.Socket{p1, p2})

		matchmaker.Accept(p1)
		reply, err := matchmaker.Accept(p2)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if reply.Method != "Game.Create" {
			t.Errorf("Expected method %v, got %v", "Game.Create", reply.Method)
		}
	})

	t.Run("deny match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch([]pkg.Socket{p1, p2})

		if _, err := matchmaker.Deny(p1); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check p1
		data, err := io.ReadAll(p1)
		if err != nil {
			t.Fatalf("Could not read data from socket: %v", err)
		}

		var response pkg.Response
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatalf("Could not parse response: %v", err)
		}
		if response.Type != pkg.MatchDeclined {
			t.Errorf("Expected type %v, got %v", pkg.MatchDeclined, response.Type)
		}

		// Check p2
		data, err = io.ReadAll(p2)
		if err != nil {
			t.Fatalf("Could not read data from socket: %v", err)
		}
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatalf("Could not parse response: %v", err)
		}
		if response.Type != pkg.MatchDeclined {
			t.Errorf("Expected type %v, got %v", pkg.MatchDeclined, response.Type)
		}
	})

	t.Run("deny not in match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()
		p1 := pkg.NewTestSocket()

		_, err := matchmaker.Deny(p1)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrMatchNotFound {
			t.Errorf("Expected error %v, got %v", pkg.ErrMatchNotFound, err)
		}
	})
}
