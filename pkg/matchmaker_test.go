package pkg_test

import (
	"reflect"
	"testing"
	"time"

	"git.internal.com/wingspan/pkg"
)

func TestMatchmaker(t *testing.T) {
	t.Run("match found", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})

		assertResponse(t, p1, pkg.MatchFound)
		assertResponse(t, p2, pkg.MatchFound)
	})

	t.Run("accept match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})
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
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})

		matchmaker.Accept(p1)
		reply, err := matchmaker.Accept(p2)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if reply.Method != "Game.Create" {
			t.Errorf("Expected method %v, got %v", "Game.Create", reply.Method)
		}

		expected := []pkg.Socket{p1, p2}
		confirmed := reply.Params.([]pkg.Socket)
		if !reflect.DeepEqual(confirmed, expected) {
			t.Errorf("Expected %v, got %v", expected, confirmed)
		}
	})

	t.Run("deny match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})

		if _, err := matchmaker.Decline(p1); err != nil {
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
		matchmaker := pkg.NewMatchmaker(time.Second)
		p1 := pkg.NewTestSocket()

		_, err := matchmaker.Decline(p1)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrMatchNotFound {
			t.Errorf("Expected error %v, got %v", pkg.ErrMatchNotFound, err)
		}
	})

	t.Run("removes match when denied", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})
		matchmaker.Decline(p2)

		_, err := matchmaker.Accept(p1)
		if err == nil {
			t.Fatal("Expected error")
		}
		if err != pkg.ErrMatchNotFound {
			t.Errorf("Expected error %v, got %v", pkg.ErrMatchNotFound, err)
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})

		p3 := pkg.NewTestSocket()
		p4 := pkg.NewTestSocket()
		matchmaker.CreateMatch(nil, []pkg.Socket{p3, p4})

		matchmaker.Decline(p2)
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
		matchmaker := pkg.NewMatchmaker(time.Second)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})
		matchmaker.Accept(p2)
		reply, _ := matchmaker.Decline(p1)
		if reply == nil {
			t.Fatal("expected reply")
		}
		if reply.Method != "Queue.Add" {
			t.Errorf("Expected method %v, got %v", "Queue.Add", reply.Method)
		}

		expected := []pkg.Socket{p2, nil}
		confirmed := reply.Params.([]pkg.Socket)
		if !reflect.DeepEqual(confirmed, expected) {
			t.Errorf("Expected %v, got %v", expected, confirmed)
		}
	})

	t.Run("declines after timeout", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Millisecond)

		players := []pkg.Socket{
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
		}

		matchmaker.CreateMatch(nil, players)
		time.Sleep(2 * time.Millisecond)

		for _, player := range players {
			response, err := player.(*pkg.TestSocket).GetResponse()
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if response.Type != pkg.MatchDeclined {
				t.Errorf("Expected response %v, got %v", pkg.MatchDeclined, response.Type)
			}
		}

		_, err := matchmaker.Decline(players[0])
		if err == nil {
			t.Error("Expected error")
		}
		if err != pkg.ErrMatchNotFound {
			t.Errorf("Expected error %v, got %v", pkg.ErrMatchNotFound, err)
		}
	})

	t.Run("create without players", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Second)
		_, err := matchmaker.CreateMatch(nil, []pkg.Socket{})

		if err == nil {
			t.Fatal("Expected error trying to create match without players")
		}
		if err != pkg.ErrNoPlayers {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrNoPlayers, err)
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(time.Microsecond)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		go matchmaker.CreateMatch(nil, []pkg.Socket{p1, p2})
		go matchmaker.Accept(p1)
		go matchmaker.Decline(p2)
	})

	t.Run("match concurrency", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		match := pkg.NewMatch([]pkg.Socket{p1, p2})

		go match.Ready()

		go match.Accept(p1)
		go match.Accept(p2)

		go match.Confirmed()
		go match.Ready()
	})
}
