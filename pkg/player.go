package pkg

type Player struct {
	food  Food
	birds map[int]*Bird
}

func NewPlayer(socket Socket) *Player {
	return &Player{
		food:  Food{},
		birds: make(map[int]*Bird),
	}
}

func (p *Player) Draw(deck Deck, qty int) error {
	cards, err := deck.Draw(qty)
	if err != nil {
		return err
	}
	for _, card := range cards {
		p.birds[card.ID] = card
	}
	return nil
}

func (p *Player) GainFood(foodType FoodType, qty int) {
	p.food.Increment(foodType, qty)
}

func (p *Player) DiscardFood(foodType FoodType, qty int) error {
	return p.food.Decrement(foodType, qty)
}

func (p *Player) KeepBirds(birdIds []int) error {
	cardsToRemove := make([]int, 0)
	for k := range p.birds {
		for _, id := range birdIds {
			if k == id {
				continue
			}
			if _, ok := p.birds[id]; !ok {
				return ErrBirdCardNotFound
			}
			cardsToRemove = append(cardsToRemove, k)
		}
	}

	for _, id := range cardsToRemove {
		delete(p.birds, id)
	}

	return nil
}

func (p *Player) GetFood() Food {
	return p.food
}

func (p *Player) GetBirdCards() []*Bird {
	cards := make([]*Bird, 0)
	for _, card := range p.birds {
		cards = append(cards, card)
	}
	return cards
}
