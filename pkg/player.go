package pkg

import "math/rand"

type Player struct {
	food  []Food
	birds map[int]*Bird
}

func NewPlayer(socket Socket) *Player {
	return &Player{
		food:  make([]Food, 0),
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

func (p *Player) GainFood(qty int) {
	for i := 0; i < qty; i++ {
		random := rand.Intn(5)
		food := Food{Type: FoodType(random)}
		p.food = append(p.food, food)
	}
}

func (p *Player) GetFood() []Food {
	return p.food
}

func (p *Player) GetBirdCards() []*Bird {
	cards := make([]*Bird, 0)
	for _, card := range p.birds {
		cards = append(cards, card)
	}
	return cards
}
