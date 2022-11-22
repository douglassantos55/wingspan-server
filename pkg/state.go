package pkg

type State interface {
	Enter(*Player) error
	Process(*Player, any) error
}

type ChooseFoodState struct {
	Qty    int
	Source FoodSupplier
}

func (s *ChooseFoodState) Enter(player *Player) error {
	player.socket.Send(Response{
		Type: ChooseFood,
		Payload: GainFood{
			Amount:    s.Qty,
			Available: s.Source.List(),
		},
	})
	return nil
}

func (s *ChooseFoodState) Process(player *Player, params any) error {
	total := 0
	chosen := params.(map[FoodType]int)

	for food, qty := range chosen {
		total += qty
		if err := s.Source.GetFood(food, qty); err != nil {
			return err
		}
	}

	if total != s.Qty {
		return ErrNotEnoughFood
	}
	for food, qty := range chosen {
		player.GainFood(food, qty)
	}

	return nil
}

type DrawCardsState struct {
	Qty    int
	Source BirdList
}

func (s *DrawCardsState) Enter(player *Player) error {
	birdIds := make([]BirdID, 0)
	for _, bird := range s.Source.Birds() {
		birdIds = append(birdIds, bird.ID)
	}

	player.socket.Send(Response{
		Type: ChooseCards,
		Payload: map[string]any{
			"qty":   s.Qty,
			"cards": birdIds,
		},
	})

	return nil
}

func (s *DrawCardsState) Process(player *Player, params any) error {
	birdIds := params.([]BirdID)
	if len(birdIds) != s.Qty {
		return ErrNotEnoughCards
	}

	for _, id := range birdIds {
		bird, err := s.Source.Get(id)
		if err != nil {
			return err
		}
		player.GainBird(bird)
	}

	return nil
}

type LayEggsState struct {
	Qty   int
	Birds []BirdID
}

func (s *LayEggsState) Enter(player *Player) error {
	player.socket.Send(Response{
		Type: ChooseBirds,
		Payload: map[string]any{
			"qty":   s.Qty,
			"birds": s.Birds,
		},
	})
	return nil
}

func (s *LayEggsState) Process(player *Player, params any) error {
	chosenBirds, ok := params.(map[BirdID]int)
	if !ok {
		return ErrUnexpectedValue
	}

	for id, qty := range chosenBirds {
		bird := player.board.GetBird(id)
		if bird == nil {
			return ErrBirdCardNotFound
		}
		if err := bird.LayEggs(qty); err != nil {
			return err
		}
	}

	return nil
}
