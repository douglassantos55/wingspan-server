package pkg_test

import (
	"reflect"
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

		response, err := p1.GetResponse()
		if err != nil {
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

		params := reply.Params.([]pkg.Socket)
		expected := []pkg.Socket{p1, p2}
		if !reflect.DeepEqual(params, expected) {
			t.Errorf("Expected %v, got %v", expected, params)
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
		response, err := p1.GetResponse()
		if err != nil {
			t.Fatalf("Could not read data from socket: %v", err)
		}
		if response.Type != pkg.MatchDeclined {
			t.Errorf("Expected type %v, got %v", pkg.MatchDeclined, response.Type)
		}

		// Check p2
		response, err = p2.GetResponse()
		if err != nil {
			t.Fatalf("Could not read data from socket: %v", err)
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

	t.Run("removes match when denied", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch([]pkg.Socket{p1, p2})
		matchmaker.Deny(p2)

		_, err := matchmaker.Accept(p1)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err != pkg.ErrMatchNotFound {
			t.Errorf("Expected error %v, got %v", pkg.ErrMatchNotFound, err)
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		matchmaker.CreateMatch([]pkg.Socket{p1, p2})

		p3 := pkg.NewTestSocket()
		p4 := pkg.NewTestSocket()
		matchmaker.CreateMatch([]pkg.Socket{p3, p4})

		matchmaker.Deny(p2)
		if _, err := matchmaker.Accept(p1); err == nil {
			t.Error("Expected error")
		}

		matchmaker.Accept(p3)
		reply, _ := matchmaker.Accept(p4)
		if reply == nil {
			t.Fatal("Expected reply")
		}
		if reply.Method != "Game.Create" {
			t.Errorf("Expected method %v, got %v", "Game.Create", reply.Method)
		}
	})

	t.Run("confirmed are requeued", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch([]pkg.Socket{p1, p2})
		matchmaker.Accept(p2)
		reply, _ := matchmaker.Deny(p1)
		if reply == nil {
			t.Fatal("expected reply")
		}
		if reply.Method != "Queue.Add" {
			t.Errorf("Expected method %v, got %v", "Queue.Add", reply.Method)
		}

		expected := []pkg.Socket{p2}
		params := reply.Params.([]pkg.Socket)
		if !reflect.DeepEqual(params, expected) {
			t.Errorf("Expected %v, got %v", expected, params)
		}
	})
}
