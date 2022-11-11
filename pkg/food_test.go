package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestBirdfeeder(t *testing.T) {
	t.Run("get food", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(5)
		err := feeder.GetFood(pkg.Rodent, 1)

		if err != nil {
			t.Fatalf("expected no error, got \"%v\"", err)
		}
		if feeder.Len() != 4 {
			t.Errorf("expected len %v, got %v", 4, feeder.Len())
		}
		if err := feeder.GetFood(pkg.Rodent, 1); err == nil {
			t.Error("should not have seed again")
		}
	})

	t.Run("refill", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(1)
		if err := feeder.GetFood(pkg.Fish, 1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		feeder.Refill()
		if feeder.Len() != 1 {
			t.Errorf("expected len %v, got %v", 1, feeder.Len())
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(1)
		food := feeder.List()
		for foodType := range food {
			go feeder.GetFood(foodType, 1)
		}

		go feeder.Len()
		go feeder.Refill()
	})
}
