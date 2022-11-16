package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestGainFoodPower(t *testing.T) {
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

	t.Run("gain food from supply", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		power := pkg.NewGainFood([]*pkg.Player{player}, 1, pkg.Rodent, nil)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not gain food: %v", err)
		}

		food := player.GetFood()
		if len(food) != 1 {
			t.Errorf("expected %v food, got %v", 1, len(food))
		}
		if food[pkg.Rodent] != 1 {
			t.Errorf("expected %v food, got %v", 1, food[pkg.Rodent])
		}
	})
}

func TestCacheFoodPower(t *testing.T) {
	t.Run("cache from supply (no source)", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewCacheFoodPower(bird, pkg.Seed, 1, nil)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not cache food: %v", err)
		}

		if bird.CachedFood != 1 {
			t.Errorf("expected %v cached food, got %v", 1, bird.CachedFood)
		}
	})

	t.Run("cache food from birdfeeder", func(t *testing.T) {
		bird := &pkg.Bird{}
		feeder := pkg.NewBirdfeeder(10)
		power := pkg.NewCacheFoodPower(bird, pkg.Rodent, 2, feeder)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not cache food: %v", err)
		}

		if bird.CachedFood != 2 {
			t.Errorf("expected %v cached food, got %v", 2, bird.CachedFood)
		}
	})

	t.Run("cache more than available from feeder", func(t *testing.T) {
		bird := &pkg.Bird{}
		feeder := pkg.NewBirdfeeder(0)
		power := pkg.NewCacheFoodPower(bird, pkg.Rodent, 2, feeder)

		if err := power.Execute(); err == nil {
			t.Error("should error")
		}
	})
}

func TestDrawCardsPower(t *testing.T) {
	t.Run("draw from deck", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewDrawCardsPower(player, 2, deck, nil)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		if deck.Len() != 8 {
			t.Errorf("expected len %v, got %v", 8, deck.Len())
		}
		if err := player.PlayBird(9); err != nil {
			t.Errorf("could not play bird: %v", err)
		}
	})

	t.Run("draw from tray", func(t *testing.T) {
		tray := pkg.NewBirdTray(20)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewDrawCardsPower(player, 2, nil, tray)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		// TODO: implement
	})
}
