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

		power := pkg.DrawFromDeck(player, 2, deck)

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
		tray := pkg.NewBirdTray(2)
		player := pkg.NewPlayer(pkg.NewTestSocket())
		power := pkg.DrawFromTray(player, 2, tray)

		tray.Refill(pkg.NewDeck(100))

		if err := power.Execute(); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}

		cards := player.GetBirdCards()
		if len(cards) != 2 {
			t.Errorf("expected %v cards, got %v", 2, len(cards))
		}
		if tray.Len() != 0 {
			t.Errorf("expected empty tray, got %v", tray.Len())
		}
	})

	t.Run("draw from tray (choosing)", func(t *testing.T) {
		tray := pkg.NewBirdTray(20)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		tray.Refill(pkg.NewDeck(100))
		power := pkg.DrawFromTray(player, 2, tray)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}
		// TODO: implement
	})
}

func TestTuckPower(t *testing.T) {
	t.Run("tuck from deck", func(t *testing.T) {
		bird := &pkg.Bird{}
		deck := pkg.NewDeck(20)
		power := pkg.TuckFromDeck(bird, 2, deck)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not tuck from hand: %v", err)
		}
		if deck.Len() != 18 {
			t.Errorf("expected len %v, got %v", 18, deck.Len())
		}
		if bird.TuckedCards != 2 {
			t.Errorf("expected %v tucked cards, got %v", 2, bird.TuckedCards)
		}
	})

	t.Run("tuck from hand", func(t *testing.T) {
		bird := &pkg.Bird{}
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1)})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(2)})

		power := pkg.TuckFromHand(bird, 2, player)
		if err := power.Execute(); err != nil {
			t.Fatalf("could not tuck from hand: %v", err)
		}

		if len(player.GetBirdCards()) != 0 {
			t.Errorf("expected len %v, got %v", 0, len(player.GetBirdCards()))
		}
		if bird.TuckedCards != 2 {
			t.Errorf("expected %v tucked cards, got %v", 2, bird.TuckedCards)
		}
	})

	t.Run("tuck from hand (choosing)", func(t *testing.T) {
		bird := &pkg.Bird{}
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1)})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(2)})

		power := pkg.TuckFromHand(bird, 1, player)
		if err := power.Execute(); err != nil {
			t.Fatalf("could not tuck from hand: %v", err)
		}

		//TODO: check for choose response and do stuff
	})
}

func TestFishingPower(t *testing.T) {
	t.Run("unsuccessfull", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewFishingPower(bird, 1, pkg.Fish)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.CachedFood != 0 {
			t.Errorf("expected %v cached food, got %v", 0, bird.CachedFood)
		}
	})

	t.Run("successfull", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewFishingPower(bird, 1, pkg.Rodent)

		if err := power.Execute(); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.CachedFood != 1 {
			t.Errorf("expected %v cached food, got %v", 1, bird.CachedFood)
		}
	})
}
