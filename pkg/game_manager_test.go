package pkg_test

import (
	"reflect"
	"strconv"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func assertResponse(t testing.TB, socket *pkg.TestSocket, expected string) pkg.Response {
	t.Helper()
	response, err := socket.GetResponse()

	if err != nil {
		t.Fatalf("Failed reading response: %v", err)
	}
	if response.Type != expected {
		t.Fatalf("Expected response %v, got %v", expected, response.Type)
	}
	return *response
}

func assertFoodQty(t testing.TB, food map[pkg.FoodType]int, expected int) {
	t.Helper()
	total := 0
	for _, amount := range food {
		total += amount
	}
	if total != expected {
		t.Errorf("expected %v food, got %v", expected, total)
	}
}

func TestGameManager(t *testing.T) {
	discardFood := func(t testing.TB, player pkg.Socket, manager *pkg.GameManager) {
		t.Helper()

		response := assertResponse(t, player.(*pkg.TestSocket), pkg.ChooseCards)

		var payload pkg.ChooseResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := []string{}
		for k := range payload.Food {
			keys = append(keys, strconv.FormatInt(int64(k), 10))
		}

		if _, err := manager.DiscardFood(player, map[string]any{keys[0]: 0}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	t.Run("creates game", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		manager.Create(nil, []pkg.Socket{p1})

		if _, err := manager.ChooseBirds(p1, []any{169}); err != nil {
			t.Errorf("Expecte no error, got %v", err)
		}

		p2 := pkg.NewTestSocket()
		if _, err := manager.ChooseBirds(p2, []any{1}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("starts with cards and food", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.ChooseResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Failed parsing payload: %v", err)
		}

		if len(payload.Birds) != pkg.INITIAL_BIRDS {
			t.Errorf("Expected %v birds, got %v", pkg.INITIAL_BIRDS, len(payload.Birds))
		}

		assertFoodQty(t, payload.Food, pkg.INITIAL_FOOD)

		seenBirds := make(map[*pkg.Bird]bool)
		for _, bird := range payload.Birds {
			if bird == nil {
				t.Fatal("Expected bird, got nil")
			}
			if _, ok := seenBirds[bird]; ok {
				t.Fatalf("Duplicated bird %v", bird)
			}
			seenBirds[bird] = true
		}
	})

	t.Run("choose initial birds", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.ChooseResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Error parsing payload: %v", err)
		}

		chosenBirds := []any{payload.Birds[0].ID, payload.Birds[4].ID}
		if _, err := manager.ChooseBirds(p1, chosenBirds); err != nil {
			t.Errorf("Error chosing birds: %v", err)
		}

		response = assertResponse(t, p1, pkg.DiscardFood)
		amountToDiscard := int(response.Payload.(float64))
		if amountToDiscard != len(chosenBirds) {
			t.Errorf("Expected %v, got %v", len(chosenBirds), amountToDiscard)
		}
	})

	t.Run("discard food", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		assertResponse(t, p1, pkg.StartTurn)
		assertResponse(t, p2, pkg.WaitTurn)
	})

	t.Run("end turn", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.EndTurn(p1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertResponse(t, p1, pkg.WaitTurn)
		assertResponse(t, p2, pkg.StartTurn)
	})

	t.Run("concurrency", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		go manager.Create(nil, []pkg.Socket{p1})
		go manager.Create(nil, []pkg.Socket{p2})

		go manager.ChooseBirds(p1, []any{0})
		go manager.ChooseBirds(p2, []any{0})

		go manager.DiscardFood(p1, map[string]any{})
		go manager.DiscardFood(p2, map[string]any{})
	})

	t.Run("game over", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.PlayCard(p1, 169); err != nil {
			t.Fatalf("could not play card: %v", err)
		}

		for i := 0; i < pkg.MAX_ROUNDS; i++ {
			for j := 0; j < (pkg.MAX_TURNS-i)*2; j++ {
				if i%2 == 0 {
					if _, err := manager.EndTurn(p1); err != nil {
						t.Fatal("could not end turn")
					}
				} else {
					if _, err := manager.EndTurn(p2); err != nil {
						t.Fatal("could not end turn")
					}
				}
			}
		}

		response := assertResponse(t, p1, pkg.GameOver)
		payload := response.Payload.(string)
		if payload != "You win" {
			t.Errorf("expected message %v, got %v", "You win", payload)
		}

		response = assertResponse(t, p2, pkg.GameOver)
		payload = response.Payload.(string)
		if payload != "You lost" {
			t.Errorf("expected message %v, got %v", "You lost", payload)
		}

		// checks if game is removed from manager
		if _, err := manager.EndTurn(p1); err != pkg.ErrGameNotFound {
			t.Errorf("expected error %v, got %v", pkg.ErrGameNotFound, err)
		}
		if _, err := manager.EndTurn(p2); err != pkg.ErrGameNotFound {
			t.Errorf("expected error %v, got %v", pkg.ErrGameNotFound, err)
		}
	})

	t.Run("round end", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		for j := 0; j < pkg.MAX_TURNS*2; j++ {
			if j%2 == 0 {
				if _, err := manager.EndTurn(p1); err != nil {
					t.Fatal("could not end turn")
				}
			} else {
				if _, err := manager.EndTurn(p2); err != nil {
					t.Fatal("could not end turn")
				}
			}
		}

		assertResponse(t, p1, pkg.RoundEnded)
		assertResponse(t, p2, pkg.RoundEnded)
	})

	t.Run("draw from deck", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.DrawFromDeck(p1); err != nil {
			t.Fatalf("could not draw from deck: %v", err)
		}

		assertResponse(t, p1, pkg.BirdsDrawn)
		assertResponse(t, p2, pkg.BirdsDrawn)
	})

	t.Run("lay eggs", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.LayEggs(p1); err != nil {
			t.Fatalf("failed laying eggs: %v", err)
		}

		response := assertResponse(t, p1, pkg.ChooseBirds)
		payload := response.Payload.(map[string]any)

		if payload["qty"].(float64) != 2 {
			t.Errorf("expected %v, got %v", 2, payload["qty"])
		}
	})

	t.Run("activate Power", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.PlayCard(p1, pkg.BirdID(169)); err != nil {
			t.Fatalf("could not play card: %v", err)
		}
		if _, err := manager.ActivatePower(p1, pkg.BirdID(999)); err == nil {
			t.Error("should not activate power of missing bird")
		}
		if _, err := manager.ActivatePower(p1, pkg.BirdID(169)); err != nil {
			t.Errorf("could not activate power: %v", err)
		}
	})

	t.Run("player info", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		game, _ := manager.GetSocketGame(p1)
		players := game.TurnOrder()

		game.Disconnect(p2)

		if _, err := manager.PlayerInfo(p1, players[1].ID.String()); err != nil {
			t.Fatalf("could not get player info: %v", err)
		}

		response := assertResponse(t, p1, pkg.PlayerInfo)

		var payload pkg.PlayerInfoPayload
		pkg.ParsePayload(response.Payload, &payload)

	tray_outer:
		for _, bird := range game.BirdTray() {
			for _, received := range payload.BirdTray {
				if received.ID == bird.ID {
					continue tray_outer
				}
			}
			t.Errorf("did not find %v", bird.ID)
		}

		if !reflect.DeepEqual(payload.BirdFeeder, game.Birdfeeder()) {
			t.Errorf("expected %v, got %v", game.Birdfeeder(), payload.BirdFeeder)
		}

	player_outer:
		for _, player := range game.TurnOrder() {
			for _, received := range payload.TurnOrder {
				if received.ID == player.ID {
					continue player_outer
				}
			}
			t.Errorf("did not find player %v", player.ID)
		}

		if payload.Current != players[0].ID {
			t.Errorf("expected %v, got %v", players[0].ID, payload.Current)
		}
		if payload.Duration != 60 {
			t.Errorf("expected %v, got %v", 60, payload.Duration)
		}
		if payload.Round != 0 {
			t.Errorf("expected %v, got %v", 0, payload.Round)
		}
		if payload.Turn != 0 {
			t.Errorf("expected %v, got %v", 0, payload.Turn)
		}
		if payload.MaxTurns != pkg.MAX_TURNS {
			t.Errorf("expected %v, got %v", pkg.MAX_TURNS, payload.MaxTurns)
		}
	})

	t.Run("player info connected socket", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		socket := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		game, _ := manager.GetSocketGame(p1)
		players := game.TurnOrder()

		if _, err := manager.PlayerInfo(socket, players[0].ID.String()); err != nil {
			t.Error("should not be able to attach to an existing socket")
		}
	})

	t.Run("player info new socket", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		socket := pkg.NewTestSocket()

		manager.Create(nil, []pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		game, _ := manager.GetSocketGame(p2)
		players := game.TurnOrder()

		game.Disconnect(p2)

		if _, err := manager.PlayerInfo(socket, players[1].ID.String()); err != nil {
			t.Fatalf("could not get player info: %v", err)
		}

		assertResponse(t, socket, pkg.PlayerInfo)

		if _, err := manager.EndTurn(socket); err != nil {
			t.Errorf("should end turn, got: %v", err)
		}

		game.Broadcast(pkg.Response{Type: pkg.MatchFound})
		assertResponse(t, socket, pkg.MatchFound)

		if res, err := p2.GetResponse(); err == nil && res.Type == pkg.MatchFound {
			t.Errorf("removed socket should not receive responses, got %v", res)
		}
	})
}
