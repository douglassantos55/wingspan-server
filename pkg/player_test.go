package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestPlayer(t *testing.T) {
	t.Run("draw", func(t *testing.T) {
		deck := pkg.NewDeck(10)

		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if err := player.Draw(deck, 5); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if deck.Len() != 5 {
			t.Errorf("Expected %v length, got %v", 5, deck.Len())
		}

		cards := player.GetBirdCards()
		if len(cards) != 5 {
			t.Errorf("Expected %v cards, got %v", 5, len(cards))
		}
	})

	t.Run("draw more than in deck", func(t *testing.T) {
		deck := pkg.NewDeck(0)

		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if err := player.Draw(deck, 5); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("gain food", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		player.GainFood(pkg.Invertebrate, 10)
		assertFoodQty(t, player.GetFood(), 10)
	})

	t.Run("keep birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		// card not found
		if err := player.KeepBirds([]pkg.BirdID{0000}); err == nil {
			t.Error("Expected error, got nothing")
		}

		// works properly
		if err := player.KeepBirds([]pkg.BirdID{9}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// cards are removed
		if err := player.KeepBirds([]pkg.BirdID{8}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("keep all birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		if err := player.KeepBirds([]pkg.BirdID{9, 8, 7, 6, 5}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("discard food not found", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Fruit, 2)
		player.GainFood(pkg.Seed, 1)
		player.GainFood(pkg.Fish, 3)

		err := player.DiscardFood(pkg.Invertebrate, 2)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrFoodNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrFoodNotFound, err)
		}
	})

	t.Run("discard food more than exists", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Fruit, 2)
		player.GainFood(pkg.Seed, 1)
		player.GainFood(pkg.Fish, 3)

		err := player.DiscardFood(pkg.Fruit, 5)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrNotEnoughFood {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrNotEnoughFood, err)
		}
	})

	t.Run("discard food", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Seed, 1)
		if err := player.DiscardFood(pkg.Seed, 1); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		// make sure food is removed
		err := player.DiscardFood(pkg.Seed, 1)
		if err == nil {
			t.Error("Expected error got nothing")
		}
		if err != pkg.ErrFoodNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrFoodNotFound, err)
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		go player.GainFood(pkg.Seed, 1)
		go player.GainFood(pkg.Fruit, 1)

		go player.DiscardFood(pkg.Seed, 1)
		go player.DiscardFood(pkg.Fruit, 1)

		deck := pkg.NewDeck(50)

		go player.Draw(deck, 1)
		go player.Draw(deck, 1)

		go player.KeepBirds([]pkg.BirdID{0})
		go player.KeepBirds([]pkg.BirdID{1})

		go player.GetFood()
		go player.GetBirdCards()
	})

	t.Run("play bird", func(t *testing.T) {
	})

	t.Run("count eggs to lay", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if player.GetEggsToLay() != 1 {
			t.Errorf("expected %v, got %v", 1, player.GetEggsToLay())
		}
	})
}
