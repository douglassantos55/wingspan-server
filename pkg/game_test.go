package pkg_test

import (
	"testing"
	"time"

	"git.internal.com/wingspan/pkg"
)

func TestGame(t *testing.T) {
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

		var payload pkg.StartingResources
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
		err := game.ChooseBirds(p2, []int{0})

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

		err := game.ChooseBirds(p1, []int{9999})
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
		if err := game.ChooseBirds(p1, []int{169}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// make sure the other cards are removed
		if err := game.ChooseBirds(p1, []int{168}); err == nil {
			t.Error("expected error, got nothing")
		}
	})

	t.Run("choice timeout", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)

		game.Start(time.Millisecond)
		time.Sleep(2 * time.Millisecond)

		assertResponse(t, p1, pkg.GameCanceled)
		assertResponse(t, p2, pkg.GameCanceled)
	})

	t.Run("discard food no game", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1}, time.Second)
		err := game.DiscardFood(p2, pkg.Fish, 1)

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

		var payload pkg.StartingResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := make([]pkg.FoodType, 0, len(payload.Food))
		for ft := range payload.Food {
			keys = append(keys, ft)
		}

		foodType := keys[0]
		if err := game.DiscardFood(p1, foodType, payload.Food[foodType]); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		foodType = keys[1]
		if err := game.DiscardFood(p1, foodType, payload.Food[foodType]+1); err == nil {
			t.Error("expected error, got nothing")
		}
	})

	t.Run("discard timeout", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)
		game.Start(time.Millisecond)

		response, _ := p1.GetResponse()
		var payload pkg.StartingResources
		pkg.ParsePayload(response.Payload, &payload)

		keys := make([]pkg.FoodType, 0, len(payload.Food))
		for ft := range payload.Food {
			keys = append(keys, ft)
		}

		if err := game.DiscardFood(p1, keys[0], 1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		time.Sleep(2 * time.Millisecond)

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

			var payload pkg.StartingResources
			pkg.ParsePayload(response.Payload, &payload)

			keys := make([]pkg.FoodType, 0, len(payload.Food))
			for ft := range payload.Food {
				keys = append(keys, ft)
			}

			if err := game.DiscardFood(player, keys[0], 1); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

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

		if err := game.DiscardFood(p1, 3, 0); err != nil {
			t.Fatalf("expected no error, got \"%v\"", err)
		}
		if err := game.DiscardFood(p2, 1, 0); err != nil {
			t.Fatalf("expected no error, got \"%v\"", err)
		}

		time.Sleep(3 * time.Millisecond)

		assertResponse(t, p1, pkg.WaitTurn)
		assertResponse(t, p2, pkg.StartTurn)
	})

	t.Run("concurrency", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{p1, p2}, time.Second)

		game.Start(time.Millisecond)

		go game.ChooseBirds(p1, []int{0})
		go game.ChooseBirds(p2, []int{0})

		go game.DiscardFood(p1, 0, 0)
		go game.DiscardFood(p2, 0, 0)

		go game.StartTurn()
		go game.EndTurn()
	})
}
