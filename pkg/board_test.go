package pkg_test

import (
	"reflect"
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

	t.Run("get bird", func(t *testing.T) {
		board := pkg.NewBoard()

		bird := &pkg.Bird{ID: pkg.BirdID(7)}
		if err := board.PlayBird(bird); err != nil {
			t.Errorf("should not have error, got %v", err)
		}

		if board.GetBird(bird.ID) == nil {
			t.Error("should find bird")
		}
	})

	t.Run("total eggs", func(t *testing.T) {
		board := pkg.NewBoard()
		board.PlayBird(&pkg.Bird{EggCount: 2})
		board.PlayBird(&pkg.Bird{EggCount: 3})

		if board.TotalEggs() != 5 {
			t.Errorf("expected %v eggs, got %v", 5, board.TotalEggs())
		}
	})

	t.Run("birds with eggs", func(t *testing.T) {
		board := pkg.NewBoard()

		board.PlayBird(&pkg.Bird{ID: pkg.BirdID(1), EggCount: 2})
		board.PlayBird(&pkg.Bird{ID: pkg.BirdID(2), EggCount: 3})

		expected := map[pkg.BirdID]int{1: 2, 2: 3}
		if !reflect.DeepEqual(expected, board.GetBirdsWithEggs()) {
			t.Errorf("Expected %v, got %v", expected, board.GetBirdsWithEggs())
		}
	})
}
