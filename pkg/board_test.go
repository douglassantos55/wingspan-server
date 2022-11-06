package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestBoard(t *testing.T) {
	t.Run("play bird", func(t *testing.T) {
		board := pkg.NewBoard()
		bird := &pkg.Bird{}

		if err := board.PlayBird(bird); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := board.PlayBird(bird); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := board.PlayBird(bird); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := board.PlayBird(bird); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := board.PlayBird(bird); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := board.PlayBird(bird); err == nil {
			t.Error("expected error, got nothing")
		}
	})

	t.Run("leftmost exposed", func(t *testing.T) {
		board := pkg.NewBoard()

		board.PlayBird(&pkg.Bird{Habitat: pkg.Forest})
		board.PlayBird(&pkg.Bird{Habitat: pkg.Forest})

		column := board.Exposed(pkg.Forest)
		if column != 2 {
			t.Errorf("Expected index %v, got %v", 2, column)
		}

		board.PlayBird(&pkg.Bird{Habitat: pkg.Forest})
		if board.Exposed(pkg.Forest) != 3 {
			t.Errorf("Expected index %v, got %v", 3, column)
		}
	})
}
