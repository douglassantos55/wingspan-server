package pkg_test

import (
	"reflect"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestBirdTray(t *testing.T) {
	t.Run("refill", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		tray := pkg.NewBirdTray(3)

		if err := tray.Refill(deck); err != nil {
			t.Errorf("expected no error, got \"%v\"", err)
		}

		curr := tray.Birds()
		tray.Get(curr[0].ID)
		tray.Get(curr[1].ID)

		tray.Refill(deck)

		if tray.Len() != 3 {
			t.Errorf("expected len %v, got %v", 3, tray.Len())
		}
		if reflect.DeepEqual(curr, tray.Birds()) {
			t.Error("expected different birds on tray, got the same")
		}
	})

	t.Run("reset", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		tray := pkg.NewBirdTray(3)

		tray.Refill(deck)
		original := tray.Birds()

		if err := tray.Reset(deck); err != nil {
			t.Fatalf("expected no error, got \"%v\"", err)
		}
		if reflect.DeepEqual(original, tray.Birds()) {
			t.Errorf("expected %v, got %v", original, tray.Birds())
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		tray := pkg.NewBirdTray(3)

		go tray.Refill(deck)
		go tray.Birds()
		go tray.Len()
		go tray.Reset(deck)
	})
}
