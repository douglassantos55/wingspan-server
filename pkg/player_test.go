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

		player.GainFood(10)

		equal := 0
		var prev pkg.Food
		for _, food := range player.GetFood() {
			if food == prev {
				equal++
			}
		}

		if equal == len(player.GetFood()) {
			t.Error("Should be random")
		}

		food := player.GetFood()
		if len(food) != 10 {
			t.Errorf("Expected %v food, got %v", 10, len(food))
		}
	})

	t.Run("keep birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		// card not found
		if err := player.KeepBirds([]int{0000}); err == nil {
			t.Error("Expected error, got nothing")
		}

		// works properly
		if err := player.KeepBirds([]int{9}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// cards are removed
		if err := player.KeepBirds([]int{8}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("keep all birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		if err := player.KeepBirds([]int{9, 8, 7, 6, 5}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}
