package pkg

import "math/rand"

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

type DrawFromDeckPower struct {
	Player *Player
	Qty    int
	Deck   Deck
}

func DrawFromDeck(player *Player, qty int, deck Deck) *DrawFromDeckPower {
	return &DrawFromDeckPower{
		Player: player,
		Qty:    qty,
		Deck:   deck,
	}
}

func (p *DrawFromDeckPower) Execute() error {
	birds, err := p.Deck.Draw(p.Qty)
	if err != nil {
		return err
	}
	for _, bird := range birds {
		p.Player.GainBird(bird)
	}
	return nil
}

type DrawFromTrayPower struct {
	Player   *Player
	Qty      int
	BirdTray *BirdTray
}

func DrawFromTray(player *Player, qty int, tray *BirdTray) *DrawFromTrayPower {
	return &DrawFromTrayPower{
		Player:   player,
		Qty:      qty,
		BirdTray: tray,
	}
}

func (p *DrawFromTrayPower) Execute() error {
	if p.BirdTray.Len() < p.Qty {
		return ErrNotEnoughCards
	}

	if p.BirdTray.Len() == p.Qty {
		for _, bird := range p.BirdTray.Birds() {
			if _, err := p.BirdTray.Get(bird.ID); err != nil {
				return err
			}
			p.Player.GainBird(bird)
		}
	} else {
		// TODO: choose cards
	}

	return nil
}

type TuckFromDeckPower struct {
	Bird *Bird
	Qty  int
	Deck Deck
}

func TuckFromDeck(bird *Bird, qty int, deck Deck) *TuckFromDeckPower {
	return &TuckFromDeckPower{
		Bird: bird,
		Qty:  qty,
		Deck: deck,
	}
}

func (p *TuckFromDeckPower) Execute() error {
	if _, err := p.Deck.Draw(p.Qty); err != nil {
		return err
	}
	p.Bird.TuckCards(p.Qty)
	return nil
}

type TuckFromHandPower struct {
	Bird   *Bird
	Qty    int
	Player *Player
}

func TuckFromHand(bird *Bird, qty int, player *Player) *TuckFromHandPower {
	return &TuckFromHandPower{
		Bird:   bird,
		Qty:    qty,
		Player: player,
	}
}

func (p *TuckFromHandPower) Execute() error {
	hand := p.Player.GetBirdCards()

	if len(hand) < p.Qty {
		return ErrNotEnoughCards
	}

	if len(hand) == p.Qty {
		if err := p.Player.KeepBirds(nil); err != nil {
			return err
		}
		p.Bird.TuckCards(p.Qty)
	} else {
		// TODO: choose birds
	}

	return nil
}

type FishingPower struct {
	Qty  int
	Food FoodType
	Bird *Bird
}

func NewFishingPower(bird *Bird, qty int, food FoodType) *FishingPower {
	return &FishingPower{
		Food: food,
		Bird: bird,
		Qty:  qty,
	}
}

func (p *FishingPower) Execute() error {
	random := rand.Intn(FOOD_TYPE_COUNT)
	if FoodType(random) == p.Food {
		p.Bird.CacheFood(p.Qty)
	}
	return nil
}

type HuntingPower struct {
	Bird *Bird
	Deck Deck
}

func NewHuntingPower(bird *Bird, deck Deck) *HuntingPower {
	return &HuntingPower{
		Bird: bird,
		Deck: deck,
	}
}

func (p *HuntingPower) Execute() error {
	birds, err := p.Deck.Draw(1)
	if err != nil {
		return err
	}
	for _, bird := range birds {
		if bird.Wingspan < p.Bird.HuntingPower {
			p.Bird.TuckCards(1)
		}
	}
	return nil
}

type LayEggsPower struct {
	Birds []*Bird
	Qty   int
	Nest  NestType
}

func NewLayEggsPower(birds []*Bird, qty int, nest NestType) *LayEggsPower {
	return &LayEggsPower{
		Birds: birds,
		Qty:   qty,
		Nest:  nest,
	}
}

func (p *LayEggsPower) Execute() error {
	if p.Nest != -1 {
		for _, bird := range p.Birds {
			if bird.NestType == p.Nest {
				if err := bird.LayEggs(p.Qty); err != nil {
					return err
				}
			}
		}
	} else {
		if len(p.Birds) == 1 {
			return p.Birds[0].LayEggs(p.Qty)
		}
		// TODO: choose bird to lay egg
	}
	return nil
}
