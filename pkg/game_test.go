package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestGame(t *testing.T) {
	t.Run("create without players", func(t *testing.T) {
		_, err := pkg.NewGame([]pkg.Socket{})

		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrNoPlayers {
			t.Errorf("EXpected error \"%v\", got \"%v\"", pkg.ErrNoPlayers, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		game, err := pkg.NewGame([]pkg.Socket{socket})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if game == nil {
			t.Error("Expected game, got nothing")
		}
	})

	t.Run("start", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		game, _ := pkg.NewGame([]pkg.Socket{socket})

		game.Start()
		response := assertResponse(t, socket, pkg.ChooseCards)

		var payload pkg.StartingResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(payload.Birds) != pkg.INITIAL_BIRDS {
			t.Errorf("expected %v cards, got %v", pkg.INITIAL_BIRDS, len(payload.Birds))
		}
		if len(payload.Food) != pkg.INITIAL_FOOD {
			t.Errorf("expected %v cards, got %v", pkg.INITIAL_FOOD, len(payload.Food))
		}
	})

	t.Run("choose no player", func(t *testing.T) {
		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		game, _ := pkg.NewGame([]pkg.Socket{p1})
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
		game, _ := pkg.NewGame([]pkg.Socket{p1})

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
		game, _ := pkg.NewGame([]pkg.Socket{p1})

		// keep just one
		if err := game.ChooseBirds(p1, []int{169}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// make sure the other cards are removed
		if err := game.ChooseBirds(p1, []int{168}); err == nil {
			t.Error("expected error, got nothing")
		}
	})
}
