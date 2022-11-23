package pkg_test

import (
	"reflect"
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

		player.GainFood(pkg.Invertebrate, 10)
		assertFoodQty(t, player.GetFood(), 10)
	})

	t.Run("keep birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		// card not found
		if err := player.KeepBirds([]pkg.BirdID{0000}); err == nil {
			t.Error("Expected error, got nothing")
		}

		// works properly
		if err := player.KeepBirds([]pkg.BirdID{9}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// cards are removed
		if err := player.KeepBirds([]pkg.BirdID{8}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("keep all birds", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.Draw(pkg.NewDeck(10), 5)

		if err := player.KeepBirds([]pkg.BirdID{9, 8, 7, 6, 5}); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("discard food not found", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Fruit, 2)
		player.GainFood(pkg.Seed, 1)
		player.GainFood(pkg.Fish, 3)

		err := player.DiscardFood(pkg.Invertebrate, 2)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrFoodNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrFoodNotFound, err)
		}
	})

	t.Run("discard food more than exists", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Fruit, 2)
		player.GainFood(pkg.Seed, 1)
		player.GainFood(pkg.Fish, 3)

		err := player.DiscardFood(pkg.Fruit, 5)
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrNotEnoughFood {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrNotEnoughFood, err)
		}
	})

	t.Run("discard food", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		player.GainFood(pkg.Seed, 1)
		if err := player.DiscardFood(pkg.Seed, 1); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		// make sure food is removed
		err := player.DiscardFood(pkg.Seed, 1)
		if err == nil {
			t.Error("Expected error got nothing")
		}
		if err != pkg.ErrFoodNotFound {
			t.Errorf("Expected error \"%v\", got \"%v\"", pkg.ErrFoodNotFound, err)
		}
	})

	t.Run("concurrency", func(t *testing.T) {
		player := pkg.NewPlayer(pkg.NewTestSocket())

		go player.GainFood(pkg.Seed, 1)
		go player.GainFood(pkg.Fruit, 1)

		go player.DiscardFood(pkg.Seed, 1)
		go player.DiscardFood(pkg.Fruit, 1)

		deck := pkg.NewDeck(50)

		go player.Draw(deck, 1)
		go player.Draw(deck, 1)

		go player.KeepBirds([]pkg.BirdID{0})
		go player.KeepBirds([]pkg.BirdID{1})

		go player.GetFood()
		go player.GetBirdCards()

		go player.TotalScore()
		go player.CountFood()
		go player.GetCardsToDraw()
		go player.GainBird(&pkg.Bird{})
		go player.GetBirdCards()

		go player.SetState(&pkg.DrawCardsState{Source: pkg.NewBirdTray(10)})
		go player.Process([]pkg.BirdID{})

		go player.GetEggCost(pkg.Forest)
		go player.GetEggsToLay()
		go player.GetFoodToGain()
	})

	t.Run("play bird", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{ID: pkg.BirdID(1)}
		player.GainBird(bird)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := player.PlayBird(2); err == nil {
			t.Error("expected error, got nothing")
		}

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("count eggs to lay", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if player.GetEggsToLay() != 2 {
			t.Errorf("expected %v, got %v", 2, player.GetEggsToLay())
		}

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1), Habitat: pkg.Grassland})
		player.PlayBird(1)

		if player.GetEggsToLay() != 2 {
			t.Errorf("expected %v, got %v", 2, player.GetEggsToLay())
		}
	})

	t.Run("lay egg", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:       pkg.BirdID(1),
			EggLimit: 2,
		}

		player.GainBird(bird)
		player.PlayBird(bird.ID)

		birdWithEggs, err := player.LayEgg(bird.ID, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if birdWithEggs != bird {
			t.Errorf("expected %v to be the same as %v", bird, birdWithEggs)
		}
	})

	t.Run("invalid lay egg", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		_, err := player.LayEgg(949, 1)

		if err == nil {
			t.Fatal("expected error, got nothing")
		}
		if err != pkg.ErrBirdCardNotFound {
			t.Errorf("expected error %v, got %v", pkg.ErrBirdCardNotFound, err)
		}
	})

	t.Run("lay egg full", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID: pkg.BirdID(1),
		}

		player.GainBird(bird)
		player.PlayBird(bird.ID)

		_, err := player.LayEgg(bird.ID, 1)
		if err == nil {
			t.Fatal("expected error, got nothing")
		}
		if err != pkg.ErrEggLimitReached {
			t.Errorf("expected error %v, got %v", pkg.ErrEggLimitReached, err)
		}
	})

	t.Run("count cards to draw", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if player.GetCardsToDraw() != 1 {
			t.Errorf("expected %v, got %v", 1, player.GetCardsToDraw())
		}

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1), Habitat: pkg.Wetland})
		player.PlayBird(1)

		if player.GetCardsToDraw() != 1 {
			t.Errorf("expected %v, got %v", 1, player.GetCardsToDraw())
		}
	})

	t.Run("pay food cost", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:            pkg.BirdID(1),
			FoodCondition: pkg.And,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish:   1,
				pkg.Rodent: 1,
			},
		}

		player.GainBird(bird)
		player.GainFood(pkg.Fish, 2)
		player.GainFood(pkg.Rodent, 2)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := player.PlayBird(2); err == nil {
			t.Error("expected error, got nothing")
		}

		expected := map[pkg.FoodType]int{
			pkg.Fish:   1,
			pkg.Rodent: 1,
		}

		if !reflect.DeepEqual(expected, player.GetFood()) {
			t.Errorf("Expected food %v, got %v", expected, player.GetFood())
		}

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("not enough food", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:            pkg.BirdID(1),
			FoodCondition: pkg.And,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish:   2,
				pkg.Rodent: 2,
			},
		}

		player.GainBird(bird)
		player.GainFood(pkg.Fish, 1)
		player.GainFood(pkg.Rodent, 2)

		if err := player.PlayBird(bird.ID); err != pkg.ErrNotEnoughFood {
			t.Errorf("expected error %v, got %v", pkg.ErrNotEnoughFood, err)
		}
		if err := player.PlayBird(2); err == nil {
			t.Error("expected error, got nothing")
		}

		expected := map[pkg.FoodType]int{
			pkg.Fish:   1,
			pkg.Rodent: 2,
		}

		if !reflect.DeepEqual(expected, player.GetFood()) {
			t.Errorf("Expected food %v, got %v", expected, player.GetFood())
		}
	})

	t.Run("food cost or unique", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:            pkg.BirdID(1),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish:   2,
				pkg.Rodent: 2,
			},
		}

		player.GainBird(bird)
		player.GainFood(pkg.Fish, 1)
		player.GainFood(pkg.Rodent, 2)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expected := map[pkg.FoodType]int{pkg.Fish: 1}

		if !reflect.DeepEqual(expected, player.GetFood()) {
			t.Errorf("Expected food %v, got %v", expected, player.GetFood())
		}

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("food cost or multiple", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:            pkg.BirdID(1),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish:   1,
				pkg.Rodent: 1,
			},
		}

		player.GainBird(bird)
		player.GainFood(pkg.Fish, 1)
		player.GainFood(pkg.Rodent, 2)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		assertResponse(t, socket, pkg.PayBirdCost)
	})

	t.Run("choose one food cost", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{
			ID:            pkg.BirdID(1),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish:   1,
				pkg.Rodent: 1,
			},
		}

		player.GainBird(bird)
		player.GainFood(pkg.Fish, 1)
		player.GainFood(pkg.Rodent, 2)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		response := assertResponse(t, socket, pkg.PayBirdCost)

		var payload pkg.AvailableResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Errorf("could not parse payload: %v", err)
		}

		if payload.BirdID != bird.ID {
			t.Errorf("expected ID %v, got %v", bird.ID, payload.BirdID)
		}

		expected := []pkg.FoodType{
			pkg.Fish,
			pkg.Rodent,
		}

	outer:
		for _, food := range expected {
			for _, received := range payload.Food {
				if food == received {
					continue outer
				}
			}
			t.Fatalf("expected food type %v", pkg.FoodType(food))
		}

		if err := player.PayBirdCost(bird.ID, []pkg.FoodType{pkg.Fish}, nil); err != nil {
			t.Errorf("error paying food cost: %v", err)
		}

		remaining := map[pkg.FoodType]int{pkg.Rodent: 2}
		if !reflect.DeepEqual(remaining, player.GetFood()) {
			t.Errorf("expected %v, got %v", remaining, player.GetFood())
		}

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("egg cost choice", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird1 := &pkg.Bird{
			ID:            pkg.BirdID(1),
			EggCount:      3,
			FoodCondition: pkg.And,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird1)
		player.GainFood(pkg.Fish, 2)
		player.PlayBird(bird1.ID)

		bird2 := &pkg.Bird{
			ID:            pkg.BirdID(2),
			EggCount:      1,
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird2)
		player.PlayBird(bird2.ID)

		bird3 := &pkg.Bird{
			ID:            pkg.BirdID(3),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird3)
		if err := player.PlayBird(bird3.ID); err != nil {
			t.Fatalf("error: %v", err)
		}

		response := assertResponse(t, socket, pkg.PayBirdCost)

		var payload pkg.AvailableResources
		pkg.ParsePayload(response.Payload, &payload)

		if len(payload.Birds) != 2 {
			t.Errorf("expected %v item, got %v", 1, len(payload.Birds))
		}
		if payload.Birds[1] != 2 {
			t.Errorf("expected %v eggs, got %v", 2, payload.Birds[1])
		}
		if payload.Birds[2] != 1 {
			t.Errorf("expected %v eggs, got %v", 1, payload.Birds[2])
		}
	})

	t.Run("egg cost direct", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird1 := &pkg.Bird{
			ID:            pkg.BirdID(1),
			EggCount:      1,
			FoodCondition: pkg.And,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird1)
		player.GainFood(pkg.Fish, 2)
		player.PlayBird(bird1.ID)

		bird2 := &pkg.Bird{
			ID:            pkg.BirdID(2),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird2)
		player.PlayBird(bird2.ID)

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("pay egg cost", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird1 := &pkg.Bird{
			ID:            pkg.BirdID(1),
			EggCount:      4,
			FoodCondition: pkg.And,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird1)
		player.GainFood(pkg.Fish, 5)
		player.PlayBird(bird1.ID)

		bird2 := &pkg.Bird{
			ID:            pkg.BirdID(2),
			EggCount:      2,
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird2)
		player.PlayBird(bird2.ID)

		bird3 := &pkg.Bird{
			ID:            pkg.BirdID(3),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird3)
		if err := player.PlayBird(bird3.ID); err != nil {
			t.Fatalf("error: %v", err)
		}

		response := assertResponse(t, socket, pkg.PayBirdCost)

		var payload pkg.AvailableResources
		pkg.ParsePayload(response.Payload, &payload)

		if err := player.PayBirdCost(payload.BirdID, nil, map[pkg.BirdID]int{1: 1}); err != nil {
			t.Fatalf("could not pay bird cost: %v", err)
		}

		assertResponse(t, socket, pkg.BirdUpdated)

		bird4 := &pkg.Bird{
			ID:            pkg.BirdID(4),
			FoodCondition: pkg.Or,
			FoodCost: map[pkg.FoodType]int{
				pkg.Fish: 1,
			},
		}

		player.GainBird(bird4)
		player.PlayBird(bird4.ID)

		response = assertResponse(t, socket, pkg.PayBirdCost)
		pkg.ParsePayload(response.Payload, &payload)

		player.PayBirdCost(payload.BirdID, nil, map[pkg.BirdID]int{1: 1, 2: 1})
		assertResponse(t, socket, pkg.BirdUpdated)

		if bird1.EggCount != 1 {
			t.Errorf("expected %v eggs, got %v", 1, bird1.EggCount)
		}
		if bird2.EggCount != 1 {
			t.Errorf("expected %v eggs, got %v", 1, bird2.EggCount)
		}
	})

	t.Run("no food cost", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		bird := &pkg.Bird{ID: pkg.BirdID(1)}
		player.GainBird(bird)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		assertResponse(t, socket, pkg.FoodUpdated)
	})

	t.Run("total score", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if player.TotalScore() != 0 {
			t.Errorf("Expected %v points, got %v", 0, player.TotalScore())
		}

		player.GainBird(&pkg.Bird{ID: pkg.BirdID(1), Points: 5, EggCount: 2})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(2), Points: 2})
		player.GainBird(&pkg.Bird{ID: pkg.BirdID(3), Points: 1})

		player.PlayBird(1)
		player.PlayBird(2)
		player.PlayBird(3)

		if player.TotalScore() != 11 {
			t.Errorf("Expected %v points, got %v", 11, player.TotalScore())
		}
	})

	t.Run("count food", func(t *testing.T) {
		socket := pkg.NewTestSocket()
		player := pkg.NewPlayer(socket)

		if player.CountFood() != 0 {
			t.Errorf("expected %v food, got %v", 0, player.CountFood())
		}

		player.GainFood(pkg.Fish, 2)
		player.GainFood(pkg.Seed, 5)
		player.GainFood(pkg.Rodent, 1)

		if player.CountFood() != 8 {
			t.Errorf("expected %v food, got %v", 8, player.CountFood())
		}
	})

	t.Run("activate when played power", func(t *testing.T) {
		bird := &pkg.Bird{Power: map[pkg.Trigger]pkg.Power{
			pkg.WhenPlayed: pkg.NewGainFood(1, pkg.Fish, nil),
		}}

		player := pkg.NewPlayer(pkg.NewTestSocket())
		player.GainBird(bird)

		if err := player.PlayBird(bird.ID); err != nil {
			t.Fatalf("could not play bird: %v", err)
		}
		if player.CountFood() != 1 {
			t.Errorf("expected %v food, got %v", 1, player.CountFood())
		}
	})
}
