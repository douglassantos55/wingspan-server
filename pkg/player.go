package pkg

import "sync"

// TODO: maybe add a state to player, like:
// choosing_cards, choosing_food, drawing_cards, playing_cards
// and block requests depending on the current state
type Player struct {
	socket Socket
	food   *sync.Map
	birds  *sync.Map
	board  *Board
}

func NewPlayer(socket Socket) *Player {
	return &Player{
		socket: socket,
		board:  NewBoard(),
		food:   new(sync.Map),
		birds:  new(sync.Map),
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

func (p *Player) LayEgg(birdId BirdID) (*Bird, error) {
	bird := p.board.GetBird(birdId)
	if bird == nil {
		return nil, ErrBirdCardNotFound
	}
	return bird, bird.LayEgg()
}

func (p *Player) GetCardsToDraw() int {
	column := p.board.Exposed(Wetland)
	return column/2 + 1
}

func (p *Player) GetEggsToLay() int {
	column := p.board.Exposed(Grassland)
	return column/2 + 2
}

func (p *Player) PlayBird(birdId BirdID) error {
	value, ok := p.birds.Load(birdId)
	if !ok {
		return ErrBirdCardNotFound
	}

	bird, ok := value.(*Bird)
	if !ok {
		return ErrUnexpectedValue
	}

	available := make(map[FoodType]int)
	for food, qty := range bird.FoodCost {
		value, ok := p.food.Load(food)
		if !ok || value.(int) < qty {
			continue
		}
		available[food] = qty
	}

	if bird.FoodCondition == Or {
		if len(available) > 1 {
			p.socket.Send(Response{
				Type: ChooseFood,
				Payload: AvailableFood{
					BirdID: bird.ID,
					Food:   available,
				},
			})
			return nil
		}
	} else {
		if len(available) != len(bird.FoodCost) {
			return ErrNotEnoughFood
		}
	}

	p.payFoodCost(available)
	return p.board.PlayBird(bird)
}

func (p *Player) PayFoodCost(birdId BirdID, foodType FoodType) error {
	_, ok := p.food.Load(foodType)
	if !ok {
		return ErrFoodNotFound
	}

	bird, ok := p.birds.Load(birdId)
	if !ok {
		return ErrBirdCardNotFound
	}

	qty := bird.(*Bird).FoodCost[foodType]
	p.payFoodCost(map[FoodType]int{foodType: qty})

	return p.board.PlayBird(bird.(*Bird))
}

func (p *Player) payFoodCost(foodCost map[FoodType]int) {
	for food, qty := range foodCost {
		value, _ := p.food.Load(food)
		if value.(int)-qty <= 0 {
			p.food.Delete(food)
		} else {
			p.food.Store(food, value.(int)-qty)
		}
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

func (p *Player) KeepBirds(birdIds []BirdID) error {
	cardsToRemove := make([]BirdID, 0)

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
