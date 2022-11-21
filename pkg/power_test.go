package pkg_test

import (
	"reflect"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestGainFoodPower(t *testing.T) {
	t.Run("gain one food from feeder", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(20)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.NewGainFood(1, pkg.Fish, feeder)

		if err := power.Execute(nil, player); err != nil {
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

		power := pkg.NewGainFood(-1, pkg.Fish, feeder)
		total, _ := feeder.GetAll(pkg.Fish)

		if err := power.Execute(nil, player); err != nil {
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
		// TODO
	})

	t.Run("gain food from supply", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		power := pkg.NewGainFood(1, pkg.Rodent, nil)

		if err := power.Execute(nil, player); err != nil {
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

	t.Run("gain any food from feeder", func(t *testing.T) {
		feeder := pkg.NewBirdfeeder(10)
		power := pkg.NewGainFood(1, -1, feeder)

		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if err := power.Execute(nil, player); err != nil {
			t.Fatalf("could not execute power: %v", err)
		}

		response := assertResponse(t, socket, pkg.ChooseFood)

		var payload pkg.GainFood
		pkg.ParsePayload(response.Payload, &payload)

		if payload.Amount != 1 {
			t.Errorf("expected qty %v, got %v", 1, payload.Amount)
		}
		if !reflect.DeepEqual(payload.Available, feeder.List()) {
			t.Errorf("expected options %v, got %v", feeder.List(), payload.Available)
		}

		keys := make([]pkg.FoodType, 0, len(payload.Available))
		for ft := range payload.Available {
			keys = append(keys, ft)
		}

		if err := player.Process(map[pkg.FoodType]int{
			keys[0]: 1,
		}); err != nil {
			t.Fatalf("could not choose food: %v", err)
		}
		if player.CountFood() != 1 {
			t.Errorf("expected %v food, got %v", 1, player.CountFood())
		}
	})
}

func TestCacheFoodPower(t *testing.T) {
	t.Run("cache from supply (no source)", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewCacheFoodPower(pkg.Seed, 1, nil)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not cache food: %v", err)
		}

		if bird.CachedFood != 1 {
			t.Errorf("expected %v cached food, got %v", 1, bird.CachedFood)
		}
	})

	t.Run("cache food from birdfeeder", func(t *testing.T) {
		bird := &pkg.Bird{}
		feeder := pkg.NewBirdfeeder(10)
		power := pkg.NewCacheFoodPower(pkg.Rodent, 2, feeder)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not cache food: %v", err)
		}

		if bird.CachedFood != 2 {
			t.Errorf("expected %v cached food, got %v", 2, bird.CachedFood)
		}
	})

	t.Run("cache more than available from feeder", func(t *testing.T) {
		bird := &pkg.Bird{}
		feeder := pkg.NewBirdfeeder(0)
		power := pkg.NewCacheFoodPower(pkg.Rodent, 2, feeder)

		if err := power.Execute(bird, nil); err == nil {
			t.Error("should error")
		}
	})
}

func TestDrawCardsPower(t *testing.T) {
	t.Run("draw from deck", func(t *testing.T) {
		deck := pkg.NewDeck(10)
		player := pkg.NewPlayer(pkg.NewTestSocket())

		power := pkg.DrawFromDeck(2, deck)

		if err := power.Execute(nil, player); err != nil {
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
		power := pkg.DrawFromTray(2, tray)

		tray.Refill(pkg.NewDeck(100))

		if err := power.Execute(nil, player); err != nil {
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
		tray := pkg.NewBirdTray(5)
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		tray.Refill(pkg.NewDeck(100))
		power := pkg.DrawFromTray(2, tray)

		if err := power.Execute(nil, player); err != nil {
			t.Fatalf("could not draw cards: %v", err)
		}

		expected := make([]pkg.BirdID, 0)
		for _, bird := range tray.Birds() {
			expected = append(expected, bird.ID)
		}

		response := assertResponse(t, socket, pkg.ChooseCards)
		payload := response.Payload.(map[string]any)

		if payload["qty"].(float64) != 2 {
			t.Errorf("expected qty %v, got %v", 2, payload["qty"])
		}

		received := payload["cards"].([]any)
		if len(received) != len(expected) {
			t.Errorf("Expected %v cards, got %v", len(expected), len(received))
		}

		if err := player.Process([]pkg.BirdID{expected[0]}); err == nil {
			t.Error("expected error, got nothing")
		}
		if err := player.Process([]pkg.BirdID{expected[0], expected[1]}); err != nil {
			t.Errorf("could not draw cards: %v", err)
		}
	})
}

func TestTuckPower(t *testing.T) {
	t.Run("tuck from deck", func(t *testing.T) {
		bird := &pkg.Bird{}
		deck := pkg.NewDeck(20)
		power := pkg.TuckFromDeck(2, deck)

		if err := power.Execute(bird, nil); err != nil {
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

		power := pkg.TuckFromHand(2)
		if err := power.Execute(bird, player); err != nil {
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
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1)})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(2)})

		power := pkg.TuckFromHand(1)
		if err := power.Execute(bird, player); err != nil {
			t.Fatalf("could not tuck from hand: %v", err)
		}

		response := assertResponse(t, socket, pkg.ChooseCards)
		payload := response.Payload.(map[string]any)

		if payload["qty"].(float64) != 1 {
			t.Errorf("expected %v, got %v", 1, payload["qty"])
		}
		if len(payload["cards"].([]any)) != 2 {
			t.Errorf("expected %v cards, got %v", 2, len(payload["cards"].([]any)))
		}

		if err := player.Process([]pkg.BirdID{2, 1}); err == nil {
			t.Error("expected error, got nothing")
		}
		if err := player.Process([]pkg.BirdID{22}); err == nil {
			t.Error("expected error, got nothing")
		}
		if err := player.Process([]pkg.BirdID{2}); err != nil {
			t.Errorf("could not choose bird: %v", err)
		}
	})
}

func TestFishingPower(t *testing.T) {
	t.Run("unsuccessfull", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewFishingPower(1, pkg.Fish)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.CachedFood != 0 {
			t.Errorf("expected %v cached food, got %v", 0, bird.CachedFood)
		}
	})

	t.Run("successfull", func(t *testing.T) {
		bird := &pkg.Bird{}
		power := pkg.NewFishingPower(1, pkg.Rodent)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.CachedFood != 1 {
			t.Errorf("expected %v cached food, got %v", 1, bird.CachedFood)
		}
	})
}

func TestHuntingPower(t *testing.T) {
	t.Run("successfull", func(t *testing.T) {
		bird := &pkg.Bird{HuntingPower: 100}
		deck := pkg.NewDeck(100)

		power := pkg.NewHuntingPower(deck)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.TuckedCards != 1 {
			t.Errorf("expected %v tucked card, got %v", 1, bird.TuckedCards)
		}
	})

	t.Run("unsuccessful", func(t *testing.T) {
		bird := &pkg.Bird{}
		deck := pkg.NewDeck(100)

		power := pkg.NewHuntingPower(deck)

		if err := power.Execute(bird, nil); err != nil {
			t.Fatalf("could not hunt: %v", err)
		}
		if bird.TuckedCards != 0 {
			t.Errorf("expected %v tucked card, got %v", 0, bird.TuckedCards)
		}
	})

	t.Run("empty deck", func(t *testing.T) {
		bird := &pkg.Bird{}
		deck := pkg.NewDeck(0)

		power := pkg.NewHuntingPower(deck)

		if err := power.Execute(bird, nil); err == nil {
			t.Fatal("should error")
		}
	})
}

func TestLayEggsPower(t *testing.T) {
	t.Run("lay on this bird", func(t *testing.T) {
		bird := &pkg.Bird{ID: pkg.BirdID(1), EggLimit: 5}

		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.GainBird(bird)

		power := pkg.NewLayEggsPower(2, -1)

		if err := power.Execute(bird, player); err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}

		if bird.EggCount != 2 {
			t.Errorf("expected %v eggs, got %v", 2, bird.EggCount)
		}
	})

	t.Run("lay on any bird", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1), EggLimit: 5})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(2), EggLimit: 5})

		power := pkg.NewLayEggsPower(1, -1)

		if err := power.Execute(nil, player); err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}

		// TODO: choose bird to lay eggs
	})

	t.Run("lay on each with nest type", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		bird1 := &pkg.Bird{ID: pkg.BirdID(1), NestType: pkg.Plataform, EggLimit: 5}
		bird2 := &pkg.Bird{ID: pkg.BirdID(2), NestType: pkg.Cavity, EggLimit: 5}
		bird3 := &pkg.Bird{ID: pkg.BirdID(3), NestType: pkg.Cavity, EggLimit: 5}

		player.GainBird(bird1)
		player.GainBird(bird2)
		player.GainBird(bird3)

		power := pkg.NewLayEggsPower(1, pkg.Cavity)

		if err := power.Execute(nil, player); err != nil {
			t.Fatalf("could not lay eggs: %v", err)
		}

		if bird1.EggCount != 0 {
			t.Errorf("expected %v egg, got %v", 0, bird2.EggCount)
		}
		if bird2.EggCount != 1 {
			t.Errorf("expected %v egg, got %v", 1, bird2.EggCount)
		}
		if bird3.EggCount != 1 {
			t.Errorf("expected %v egg, got %v", 1, bird3.EggCount)
		}
	})
}
