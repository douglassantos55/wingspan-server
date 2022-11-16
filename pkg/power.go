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
