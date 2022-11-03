package pkg

import "sync"

type Player struct {
	food  *sync.Map
	birds *sync.Map
}

func NewPlayer(socket Socket) *Player {
	return &Player{
		food:  new(sync.Map),
		birds: new(sync.Map),
	}
}

func (p *Player) Draw(deck Deck, qty int) error {
	cards, err := deck.Draw(qty)
	if err != nil {
		return err
	}
	for _, card := range cards {
		p.birds.Store(card.ID, card)
	}
	return nil
}

func (p *Player) GainBird(bird *Bird) {
	if _, loaded := p.birds.LoadOrStore(bird.ID, bird); loaded {
		panic("player already has this bird card")
	}
}

func (p *Player) GainFood(foodType FoodType, qty int) {
	if actual, ok := p.food.LoadOrStore(foodType, qty); ok {
		p.food.Store(foodType, actual.(int)+qty)
	}
}

func (p *Player) DiscardFood(foodType FoodType, qty int) error {
	actual, ok := p.food.Load(foodType)
	if !ok {
		return ErrFoodNotFound
	}

	if actual.(int) < qty {
		return ErrNotEnoughFood
	}

	v := actual.(int) - qty
	if v <= 0 {
		p.food.Delete(foodType)
	} else {
		p.food.Store(foodType, v)
	}

	return nil
}

func (p *Player) KeepBirds(birdIds []int) error {
	cardsToRemove := make([]int, 0)

	for _, bird := range p.GetBirdCards() {
		for _, id := range birdIds {
			if bird.ID == id {
				continue
			}
			if _, ok := p.birds.Load(id); !ok {
				return ErrBirdCardNotFound
			}
			cardsToRemove = append(cardsToRemove, bird.ID)
		}
	}

	for _, id := range cardsToRemove {
		p.birds.Delete(id)
	}

	return nil
}

func (p *Player) GetFood() map[FoodType]int {
	food := make(map[FoodType]int)

	p.food.Range(func(key, value any) bool {
		food[key.(FoodType)] = value.(int)
		return true
	})

	return food
}

func (p *Player) GetBirdCards() []*Bird {
	cards := make([]*Bird, 0)
	p.birds.Range(func(_, value any) bool {
		cards = append(cards, value.(*Bird))
		return true
	})
	return cards
}
