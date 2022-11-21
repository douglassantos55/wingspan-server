package pkg_test

import (
	"reflect"
	"testing"
	"time"

	"git.internal.com/wingspan/pkg"
)

func TestGame(t *testing.T) {
	discardFood := func(t testing.TB, player pkg.Socket, game *pkg.Game) {
		t.Helper()

		response := assertResponse(t, player.(*pkg.TestSocket), pkg.ChooseCards)

		var payload pkg.ChooseResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := []pkg.FoodType{}
		for k := range payload.Food {
			keys = append(keys, k)
		}

		if _, err := game.DiscardFood(player, keys[0], 0); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	t.Run("create without players", func(t *testing.T) {
		_, err := pkg.NewGame([]pkg.Socket{}, time.Second)

		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrNoPlayers {
			t.Errorf("EXpected error \"%v\", got \"%v\"", pkg.ErrNoPlayers, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		game, err := pkg.NewGame([]pkg.Socket{socket}, time.Second)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if game == nil {
			t.Error("Expected game, got nothing")
		}
	})

	t.Run("start", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{socket}, time.Second)

		game.Start(time.Second)
		response := assertResponse(t, socket, pkg.ChooseCards)

		var payload pkg.ChooseResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(payload.Birds) != pkg.INITIAL_BIRDS {
			t.Errorf("expected %v cards, got %v", pkg.INITIAL_BIRDS, len(payload.Birds))
		}

		assertFoodQty(t, payload.Food, pkg.INITIAL_FOOD)
	})

	t.Run("choose no player", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)
		err := game.ChooseBirds(p2, []pkg.BirdID{0})

		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrGameNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrGameNotFound, err)
		}
	})

	t.Run("choose invalid cards", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)

		err := game.ChooseBirds(p1, []pkg.BirdID{9999})
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrBirdCardNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrBirdCardNotFound, err)
		}
	})

	t.Run("choose cards", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)

		// keep just one
		if err := game.ChooseBirds(p1, []pkg.BirdID{169}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// make sure the other cards are removed
		if err := game.ChooseBirds(p1, []pkg.BirdID{168}); err == nil {
			t.Error("expected error, got nothing")
		}
	})

	t.Run("choice timeout", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)

		game.Start(time.Millisecond)
		time.Sleep(3 * time.Millisecond)

		assertResponse(t, p1, pkg.GameCanceled)
		assertResponse(t, p2, pkg.GameCanceled)
	})

	t.Run("discard food no game", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)
		_, err := game.DiscardFood(p2, pkg.Fish, 1)

		if err == nil {
			t.Fatal("expected error, got nothing")
		}
		if err != pkg.ErrGameNotFound {
			t.Errorf("expected error \"%v\", got \"%v\"", pkg.ErrGameNotFound, err)
		}
	})

	t.Run("discard food", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)
		game.Start(time.Second)

		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.ChooseResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := make([]pkg.FoodType, 0, len(payload.Food))
		for ft := range payload.Food {
			keys = append(keys, ft)
		}

		foodType := keys[0]
		if _, err := game.DiscardFood(p1, foodType, payload.Food[foodType]); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		foodType = keys[1]
		if _, err := game.DiscardFood(p1, foodType, payload.Food[foodType]+1); err == nil {
			t.Error("expected error, got nothing")
		}
	})

	t.Run("discard timeout", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(2 * time.Millisecond)

		response, _ := p1.GetResponse()
		var payload pkg.ChooseResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := make([]pkg.FoodType, 0, len(payload.Food))
		for ft := range payload.Food {
			keys = append(keys, ft)
		}

		if _, err := game.DiscardFood(p1, keys[0], 1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		time.Sleep(3 * time.Millisecond)

		assertResponse(t, p1, pkg.GameCanceled)
		assertResponse(t, p2, pkg.GameCanceled)
	})

	t.Run("discard turn start", func(t *testing.T) {
		players := []pkg.Socket{
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
		}

		game, _ := pkg.NewGame(players, time.Minute)
		game.Start(time.Minute)

		// Discard food for both players
		for _, player := range players {
			response, _ := player.(*pkg.TestSocket).GetResponse()

			var payload pkg.ChooseResources
			pkg.ParsePayload(response.Payload, &payload)

			keys := make([]pkg.FoodType, 0, len(payload.Food))
			for ft := range payload.Food {
				keys = append(keys, ft)
			}

			if _, err := game.DiscardFood(player, keys[0], 1); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		game.StartTurn()

		// check turn responses
		for i, player := range players {
			if i == 0 {
				assertResponse(t, player.(*pkg.TestSocket), pkg.StartTurn)
			} else {
				assertResponse(t, player.(*pkg.TestSocket), pkg.WaitTurn)
			}
		}
	})

	t.Run("start turn no player ready", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)

		if err := game.StartTurn(); err == nil {
			t.Error("Expected error got nothing")
		}
	})

	t.Run("turn timeout", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, 2*time.Millisecond)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		game.StartTurn()
		time.Sleep(3 * time.Millisecond)

		assertResponse(t, p1, pkg.WaitTurn)
		assertResponse(t, p2, pkg.StartTurn)
	})

	t.Run("round end after turns", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.StartRound(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		for i := 0; i < 2*pkg.MAX_TURNS; i++ {
			if i%2 == 0 {
				assertResponse(t, p1, pkg.StartTurn)
				assertResponse(t, p2, pkg.WaitTurn)
			} else {
				assertResponse(t, p1, pkg.WaitTurn)
				assertResponse(t, p2, pkg.StartTurn)
			}

			err := game.EndTurn()
			if i == 2*pkg.MAX_TURNS-1 && err != pkg.ErrRoundEnded {
				t.Errorf("expected error %v, got %v", pkg.RoundEnded, err)
			}
		}

		// verify that the first player changes for the next round
		assertResponse(t, p1, pkg.WaitTurn)
		assertResponse(t, p2, pkg.StartTurn)
	})

	t.Run("game ends after rounds", func(t *testing.T) {
		players := []pkg.Socket{
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
			pkg.NewTestSocket(),
		}

		game, _ := pkg.NewGame(players, time.Second)
		game.Start(time.Second)

		for _, player := range players {
			response := assertResponse(t, player.(*pkg.TestSocket), pkg.ChooseCards)

			var payload pkg.ChooseResources
			pkg.ParsePayload(response.Payload, &payload)

			keys := []pkg.FoodType{}
			for k := range payload.Food {
				keys = append(keys, k)
			}

			if _, err := game.DiscardFood(player, keys[0], 0); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		game.StartRound()

		for i := 0; i < pkg.MAX_ROUNDS; i++ {
			for j := 0; j < (pkg.MAX_TURNS-i)*3; j++ {
				err := game.EndTurn()
				if i == pkg.MAX_ROUNDS-1 && j == (pkg.MAX_TURNS-i)*3-1 && err != pkg.ErrGameOver {
					t.Errorf("expected error %v, got %v", pkg.GameOver, err)
				}
			}
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)

		game.Start(time.Millisecond)

		go game.ChooseBirds(p1, []pkg.BirdID{0})
		go game.ChooseBirds(p2, []pkg.BirdID{0})

		go game.DiscardFood(p1, 0, 0)
		go game.DiscardFood(p2, 0, 0)

		go game.StartTurn()
		go game.EndTurn()
	})

	t.Run("resets bird tray when round ends", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		original := game.BirdTray()

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.EndRound(); err != pkg.ErrRoundEnded {
			t.Fatalf("Expected error \"%v\", got \"%v\"", pkg.ErrRoundEnded, err)
		}

		if reflect.DeepEqual(original, game.BirdTray()) {
			t.Errorf("expected %v, got %v", original, game.BirdTray())
		}
	})

	t.Run("refill bird tray when turn ends", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		original := game.BirdTray()
		if err := game.DrawCards(p1); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		if err := game.DrawFromTray(p1, []pkg.BirdID{original[0].ID}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if err := game.DrawCards(p1); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		if err := game.DrawFromTray(p1, []pkg.BirdID{original[0].ID, original[1].ID}); err != pkg.ErrNotEnoughCards {
			t.Fatalf("expected error %v, got %v", pkg.ErrNotEnoughCards, err)
		}

		game.EndTurn()

		if reflect.DeepEqual(game.BirdTray(), original) {
			t.Error("should change birds in tray")
		}
		if len(game.BirdTray()) != pkg.MAX_BIRDS_TRAY {
			t.Errorf("expected %v, got %v", pkg.MAX_BIRDS_TRAY, len(game.BirdTray()))
		}
	})

	t.Run("gain food", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.GainFood(p1); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}
		response := assertResponse(t, p1, pkg.ChooseFood)

		var payload pkg.GainFood
		pkg.ParsePayload(response.Payload, &payload)

		if payload.Amount != 1 {
			t.Errorf("expected amount %v, got %v", 1, payload.Amount)
		}
		if !reflect.DeepEqual(payload.Available, game.Birdfeeder()) {
			t.Errorf("expected available %v, got %v", game.Birdfeeder(), payload.Available)
		}

		if err := game.PlayBird(p1, pkg.BirdID(167)); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}
		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{167: 2}); err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}
		if err := game.PlayBird(p1, pkg.BirdID(169)); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}

		if err := game.GainFood(p1); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		response = assertResponse(t, p1, pkg.ChooseFood)
		pkg.ParsePayload(response.Payload, &payload)

		if payload.Amount != 2 {
			t.Errorf("expected amount %v, got %v", 2, payload.Amount)
		}
	})

	t.Run("choose food", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.PlayBird(p1, pkg.BirdID(167)); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}
		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{167: 2}); err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}
		if err := game.PlayBird(p1, pkg.BirdID(169)); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}

		if err := game.GainFood(p1); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		response := assertResponse(t, p1, pkg.ChooseFood)

		var payload pkg.GainFood
		pkg.ParsePayload(response.Payload, &payload)

		if payload.Amount != 2 {
			t.Errorf("expected amount %v, got %v", 2, payload.Amount)
		}

		keys := make([]pkg.FoodType, 0, len(payload.Available))
		for ft := range payload.Available {
			keys = append(keys, ft)
		}

		expected := map[pkg.FoodType]int{keys[0]: 1, keys[1]: 1}
		if err := game.ChooseFood(p1, expected); err != nil {
			t.Errorf("could not choosed food: %v", err)
		}

		response = assertResponse(t, p1, pkg.FoodGained)

		var playerFood map[pkg.FoodType]int
		pkg.ParsePayload(response.Payload, &playerFood)

		if playerFood[keys[0]] < 1 {
			t.Errorf("expected at least %v of food type %v, got %v", 1, keys[0], playerFood[keys[0]])
		}
		if playerFood[keys[1]] < 1 {
			t.Errorf("expected at least %v of food type %v, got %v", 1, keys[1], playerFood[keys[1]])
		}
	})

	t.Run("draw from tray", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		birds := game.BirdTray()

		if err := game.DrawCards(p1); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		if err := game.DrawFromTray(p1, []pkg.BirdID{birds[0].ID}); err != nil {
			t.Errorf("expected no error, got \"%+v\"", err)
		}

		response := assertResponse(t, p1, pkg.BirdsDrawn)

		var payload []*pkg.Bird
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("could not parse payload: %v", err)
		}

		if len(payload) != 6 {
			t.Errorf("expected len %v, got %v", 6, len(payload))
		}

		response = assertResponse(t, p2, pkg.BirdsDrawn)
		if response.Payload.(float64) != 1 {
			t.Errorf("expected len %v, got %v", 1, len(payload))
		}

		if err := game.DrawCards(p1); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		if err := game.DrawFromTray(p1, []pkg.BirdID{birds[1].ID, birds[2].ID}); err != pkg.ErrNotEnoughCards {
			t.Errorf("Expected error %v, got %v", pkg.ErrNotEnoughCards, err)
		}
	})

	t.Run("draw from deck", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.DrawFromDeck(p1); err != nil {
			t.Fatalf("could not draw from deck: %v", err)
		}

		response := assertResponse(t, p1, pkg.BirdsDrawn)

		var payload []*pkg.Bird
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("could not parse payload: %v", err)
		}

		if len(payload) != 1 {
			t.Errorf("expected len %v, got %v", 1, len(payload))
		}

		response = assertResponse(t, p2, pkg.BirdsDrawn)
		if response.Payload.(float64) != 1 {
			t.Errorf("expected len %v, got %v", 1, len(payload))
		}
	})

	t.Run("play bird", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.PlayBird(p1, 169); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if err := game.PlayBird(p1, 4999); err == nil {
			t.Error("Expected error, got nothing")
		}

		assertResponse(t, p1, pkg.BoardUpdated)

		game.EndTurn()

		if err := game.PlayBird(p2, 162); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if err := game.PlayBird(p2, 4999); err == nil {
			t.Error("Expected error, got nothing")
		}

		assertResponse(t, p2, pkg.BoardUpdated)
	})

	t.Run("lay egg on bird", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{9999: 1}); err != pkg.ErrBirdCardNotFound {
			t.Errorf("expected error %v, got %v", pkg.ErrBirdCardNotFound, err)
		}

		if err := game.PlayBird(p1, 165); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}

		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{165: 2}); err != nil {
			t.Fatalf("could not lay eggs on bird: %v", err)
		}

		if err := game.PlayBird(p1, 169); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}

		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{169: 1}); err != pkg.ErrEggLimitReached {
			t.Fatalf("expected error %v, got %v", pkg.ErrEggLimitReached, err)
		}

		qty, err := game.LayEggs(p1)
		if err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}
		if qty != 2 {
			t.Errorf("expected qty %v, got %v", 2, qty)
		}

		if err := game.LayEggsOnBirds(p1, map[pkg.BirdID]int{165: 2}); err != nil {
			t.Fatalf("could not lay eggs on bird: %v", err)
		}

		response := assertResponse(t, p1, pkg.BirdUpdated)

		var payload []*pkg.Bird
		pkg.ParsePayload(response.Payload, &payload)

		if len(payload) != 1 {
			t.Errorf("expected len %v, got %v", 1, len(payload))
		}
		if payload[0].EggCount != 3 {
			t.Errorf("expected %v eggs, got %v", 3, payload[0].EggCount)
		}
	})

	t.Run("draw cards", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Second)

		discardFood(t, p1, game)
		discardFood(t, p2, game)

		err := game.DrawCards(p1)
		if err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
	})
}
