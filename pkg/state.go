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
	Source *BirdTray
}

func (s *DrawCardsState) Enter(player *Player) error {
	player.socket.Send(Response{
		Type: ChooseCards,
		Payload: map[string]any{
			"qty":   s.Qty,
			"cards": s.Source.Birds(),
		},
	})
	return nil
}

func (s *DrawCardsState) Process(player *Player, params any) error {
	birdIds := params.([]BirdID)
	if len(birdIds) != s.Qty {
		return ErrUnexpectedValue
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
