package pkg_test

import (
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
		t.Errorf("Expected response %v, got %v", expected, response.Type)
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

		var payload pkg.StartingResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := []pkg.FoodType{}
		for k := range payload.Food {
			keys = append(keys, k)
		}

		if _, err := manager.DiscardFood(player, keys[0], 0); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	t.Run("creates game", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		manager.Create([]pkg.Socket{p1})

		if _, err := manager.ChooseBirds(p1, []int{169}); err != nil {
			t.Errorf("Expecte no error, got %v", err)
		}

		p2 := pkg.NewTestSocket()
		if _, err := manager.ChooseBirds(p2, []int{1}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("starts with cards and food", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.StartingResources
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

		manager.Create([]pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.StartingResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Error parsing payload: %v", err)
		}

		chosenBirds := []int{payload.Birds[0].ID, payload.Birds[4].ID}
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

		manager.Create([]pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		assertResponse(t, p1, pkg.StartTurn)
		assertResponse(t, p2, pkg.WaitTurn)
	})

	t.Run("end turn", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})

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

		go manager.Create([]pkg.Socket{p1})
		go manager.Create([]pkg.Socket{p2})

		go manager.ChooseBirds(p1, []int{0})
		go manager.ChooseBirds(p2, []int{0})

		go manager.DiscardFood(p1, pkg.Invertebrate, 0)
		go manager.DiscardFood(p2, pkg.Invertebrate, 0)
	})

	t.Run("game over", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})

		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

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

		assertResponse(t, p1, pkg.GameOver)
		assertResponse(t, p2, pkg.GameOver)

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

		manager.Create([]pkg.Socket{p1, p2})

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

		manager.Create([]pkg.Socket{p1, p2})

		if _, err := manager.DrawFromDeck(p1, 2); err != nil {
			t.Fatalf("could not draw from deck: %v", err)
		}

		assertResponse(t, p1, pkg.BirdsDrawn)
		assertResponse(t, p2, pkg.BirdsDrawn)
	})

	t.Run("lay eggs", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})
		discardFood(t, p1, manager)
		discardFood(t, p2, manager)

		if _, err := manager.LayEggs(p1); err != nil {
			t.Fatalf("failed laying eggs: %v", err)
		}

		response := assertResponse(t, p1, pkg.SelectBirds)
		if response.Payload.(float64) != 1 {
			t.Errorf("expected %v, got %v", 1, response.Payload.(float64))
		}
	})
}
