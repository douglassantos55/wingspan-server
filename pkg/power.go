package pkg

type Power interface {
	Execute() error
}

type FoodSupplier interface {
	GetAll(FoodType) (int, error)
	GetFood(FoodType, int) error
}

type GainFoodPower struct {
	Players  []*Player
	Qty      int
	FoodType FoodType
	Source   FoodSupplier
}

func NewGainFood(players []*Player, qty int, foodType FoodType, source FoodSupplier) *GainFoodPower {
	return &GainFoodPower{
		Players:  players,
		Qty:      qty,
		FoodType: foodType,
		Source:   source,
	}

}

func (p *GainFoodPower) Execute() error {
	for _, player := range p.Players {
		if p.FoodType == -1 {
			// TODO: change state to choosing?
			break
		}

		if p.Source != nil {
			if p.Qty == -1 {
				qty, err := p.Source.GetAll(p.FoodType)
				if err != nil {
					return err
				}
				p.Qty = qty
			}
			if err := p.Source.GetFood(p.FoodType, p.Qty); err != nil {
				return err
			}
		}
		player.GainFood(p.FoodType, p.Qty)
	}
	return nil
}

type CacheFoodPower struct {
	Bird   *Bird
	Food   FoodType
	Qty    int
	Source FoodSupplier
}

func NewCacheFoodPower(bird *Bird, food FoodType, qty int, source FoodSupplier) *CacheFoodPower {
	return &CacheFoodPower{
		Bird:   bird,
		Qty:    qty,
		Food:   food,
		Source: source,
	}
}

func (p *CacheFoodPower) Execute() error {
	if p.Source != nil {
		if err := p.Source.GetFood(p.Food, p.Qty); err != nil {
			return err
		}
	}
	p.Bird.CacheFood(p.Qty)
	return nil
}

type DrawCardsPower struct {
	Player   *Player
	Qty      int
	Deck     Deck
	BirdTray *BirdTray
}

func NewDrawCardsPower(player *Player, qty int, deck Deck, tray *BirdTray) *DrawCardsPower {
	return &DrawCardsPower{
		Player:   player,
		Qty:      qty,
		BirdTray: tray,
		Deck:     deck,
	}
}

func (p *DrawCardsPower) Execute() error {
	if p.Deck != nil {
		birds, err := p.Deck.Draw(p.Qty)
		if err != nil {
			return err
		}
		for _, bird := range birds {
			p.Player.GainBird(bird)
		}
	}
	if p.BirdTray != nil {
		// TODO: choose cards
	}
	return nil
}
