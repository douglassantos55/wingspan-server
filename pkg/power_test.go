package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestPower(t *testing.T) {
	t.Run("gain one food from feeder", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(10)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewGainFood([]*pkg.Player{player}, 1, pkg.Fish, feeder)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		food := player.GetFood()

		if len(food) != 1 {
			t.Errorf("expected %v food, got %v", 1, len(food))
		}
		if food[pkg.Fish] != 1 {
			t.Errorf("expected %v food, got %v", 1, food[pkg.Fish])
		}
	})

	t.Run("gain all food from feeder", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(10)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewGainFood([]*pkg.Player{player}, -1, pkg.Fish, feeder)
		total, _ := feeder.GetAll(pkg.Fish)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		food := player.GetFood()

		if len(food) != 1 {
			t.Errorf("expected %v food, got %v", 1, len(food))
		}
		if food[pkg.Fish] != total {
			t.Errorf("expected %v food, got %v", 1, food[pkg.Fish])
		}
	})

	t.Run("all players gain 1 food from feeder", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(10)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewGainFood([]*pkg.Player{player}, 1, -1, feeder)
		if err := power.Execute(); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		// TODO: implement it
	})
}
