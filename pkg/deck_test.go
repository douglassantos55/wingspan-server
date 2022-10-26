package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestDeck(t *testing.T) {
	t.Run("create with size", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		if deck.Len() != 10 {
			t.Errorf("Expected size %v, got %v", 10, deck.Len())
		}
	})

	t.Run("draw", func(t *testing.T) {
		deck := pkg.NewDeck(10)

		initialLen := deck.Len()
		bird, err := deck.Draw(1)

		if err != nil {
			t.Fatalf("Expected bird, got error %v", err)
		}
		if bird == nil {
			t.Error("Expected bird, got nothing")
		}
		if deck.Len() != initialLen-1 {
			t.Errorf("Expected length %v, got %v", initialLen-1, deck.Len())
		}
	})

	t.Run("draw empty", func(t *testing.T) {
		deck := pkg.NewDeck(0)
		_, err := deck.Draw(1)

		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrNotEnoughCards {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrNotEnoughCards, err)
		}
	})
}
