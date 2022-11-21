package pkg

import "sync"

type BirdHand struct {
	birds *sync.Map
}

func NewBirdHand() *BirdHand {
	return &BirdHand{
		birds: new(sync.Map),
	}
}

func (h *BirdHand) Add(bird *Bird) {
	h.birds.LoadOrStore(bird.ID, bird)
}

func (h *BirdHand) Delete(id BirdID) {
	h.birds.Delete(id)
}

func (h *BirdHand) Find(id BirdID) (*Bird, bool) {
	value, _ := h.birds.Load(id)
	bird, ok := value.(*Bird)
	if !ok {
		return nil, ok
	}
	return bird, ok
}

func (h *BirdHand) Get(id BirdID) (*Bird, error) {
	value, loaded := h.birds.LoadAndDelete(id)
	if !loaded {
		return nil, ErrBirdCardNotFound
	}
	return value.(*Bird), nil
}

func (h *BirdHand) Birds() []*Bird {
	birds := make([]*Bird, 0)
	h.birds.Range(func(_, value any) bool {
		birds = append(birds, value.(*Bird))
		return true
	})
	return birds
}

type Player struct {
	socket Socket
	state  State
	food   *sync.Map
	birds  *BirdHand
	board  *Board
}

func NewPlayer(socket Socket) *Player {
	return &Player{
		socket: socket,
		board:  NewBoard(),
		food:   new(sync.Map),
		birds:  NewBirdHand(),
	}
}

func (p *Player) Draw(deck Deck, qty int) error {
	cards, err := deck.Draw(qty)
	if err != nil {
		return err
	}
	for _, card := range cards {
		p.birds.Add(card)
	}
	return nil
}

func (p *Player) GainBird(bird *Bird) {
	p.birds.Add(bird)
}

func (p *Player) GainFood(foodType FoodType, qty int) {
	if actual, ok := p.food.LoadOrStore(foodType, qty); ok {
		p.food.Store(foodType, actual.(int)+qty)
	}
}

func (p *Player) LayEgg(birdId BirdID, qty int) (*Bird, error) {
	bird := p.board.GetBird(birdId)
	if bird == nil {
		return nil, ErrBirdCardNotFound
	}
	return bird, bird.LayEggs(qty)
}

func (p *Player) GetCardsToDraw() int {
	column := p.board.Exposed(Wetland)
	return column/2 + 1
}

func (p *Player) GetEggsToLay() int {
	column := p.board.Exposed(Grassland)
	return column/2 + 2
}

func (p *Player) GetFoodToGain() int {
	column := p.board.Exposed(Forest)
	return column/2 + 1
}

func (p *Player) GetEggCost(habitat Habitat) int {
	column := p.board.Exposed(habitat)
	return (column / 2) + (column % 2)
}

func (p *Player) PlayBird(birdId BirdID) error {
	bird, ok := p.birds.Find(birdId)
	if !ok {
		return ErrBirdCardNotFound
	}

	available := make([]FoodType, 0)
	for food, qty := range bird.FoodCost {
		value, ok := p.food.Load(food)
		if !ok || value.(int) < qty {
			continue
		}
		available = append(available, food)
	}

	eggCost := p.GetEggCost(bird.Habitat)
	if p.board.TotalEggs() < eggCost {
		return ErrNotEnoughEggs
	}

	payload := AvailableResources{BirdID: bird.ID}

	birdsWithEggs := p.board.GetBirdsWithEggs()
	if eggCost > 0 && p.board.TotalEggs() > eggCost && len(birdsWithEggs) > 1 {
		payload.EggCost = eggCost
		payload.Birds = birdsWithEggs
	}

	if bird.FoodCondition == Or {
		if len(available) > 1 {
			payload.Food = available
		}
	} else {
		if len(available) != len(bird.FoodCost) {
			return ErrNotEnoughFood
		}
	}

	if payload.Food != nil || payload.EggCost != 0 {
		_, err := p.socket.Send(Response{
			Type:    PayBirdCost,
			Payload: payload,
		})
		return err
	}

	eggs := make(map[BirdID]int)
	for id := range birdsWithEggs {
		eggs[id] = eggCost
	}

	return p.PayBirdCost(birdId, available, eggs)
}

func (p *Player) PayBirdCost(birdID BirdID, food []FoodType, eggs map[BirdID]int) error {
	bird, ok := p.birds.Find(birdID)
	if !ok {
		return ErrBirdCardNotFound
	}

	if eggs != nil {
		if err := p.PayEggCost(p.GetEggCost(bird.Habitat), eggs); err != nil {
			return err
		}
	}
	if food != nil {
		if err := p.PayFoodCost(bird.FoodCost, food); err != nil {
			return err
		}
	}

	if err := p.board.PlayBird(bird); err != nil {
		return err
	}

	p.socket.Send(Response{
		Type:    BoardUpdated,
		Payload: p.board,
	})

	if err := bird.CastPower(WhenPlayed, p); err != nil {
		return err
	}

	if err := p.board.ActivatePowers(bird.Habitat, p); err != nil {
		return err
	}

	return nil
}

func (p *Player) PayEggCost(cost int, chosenEggs map[BirdID]int) error {
	total := 0
	birds := make(map[*Bird]int)

	for id, qty := range chosenEggs {
		total += qty
		bird := p.board.GetBird(id)
		birds[bird] = qty
	}

	if total != cost {
		return ErrNotEnoughEggs
	}

	for bird, qty := range birds {
		bird.EggCount -= qty
	}

	return nil
}

func (p *Player) PayFoodCost(cost map[FoodType]int, chosen []FoodType) error {
	for _, food := range chosen {
		value, ok := p.food.Load(food)
		if !ok {
			return ErrFoodNotFound
		}

		qty := cost[food]
		if value.(int)-qty <= 0 {
			p.food.Delete(food)
		} else {
			p.food.Store(food, value.(int)-qty)
		}
	}

	p.socket.Send(Response{
		Type:    FoodUpdated,
		Payload: p.GetFood(),
	})

	return nil
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

outer:
	for _, bird := range p.GetBirdCards() {
		for _, id := range birdIds {
			if bird.ID == id {
				continue outer
			}
			if _, ok := p.birds.Find(id); !ok {
				return ErrBirdCardNotFound
			}
		}
		cardsToRemove = append(cardsToRemove, bird.ID)
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

func (p *Player) CountFood() int {
	total := 0
	for _, qty := range p.GetFood() {
		total += qty
	}
	return total
}

func (p *Player) GetBirdCards() []*Bird {
	return p.birds.Birds()
}

func (p *Player) TotalScore() int {
	birds := p.board.GetBirds()
	total := len(birds)
	for _, bird := range birds {
		total += bird.Points + bird.EggCount + bird.CachedFood + bird.TuckedCards
	}
	return total
}

func (p *Player) SetState(state State) error {
	if err := state.Enter(p); err != nil {
		return err
	}
	p.state = state
	return nil
}

func (p *Player) Process(params any) error {
	if err := p.state.Process(p, params); err != nil {
		return err
	}
	p.state = nil
	return nil
}
