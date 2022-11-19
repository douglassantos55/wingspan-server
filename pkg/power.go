package pkg

import "math/rand"

type Power interface {
	Execute(*Bird, *Player) error
}

type FoodSupplier interface {
	GetAll(FoodType) (int, error)
	GetFood(FoodType, int) error
}

type GainFoodPower struct {
	Qty      int
	FoodType FoodType
	Source   FoodSupplier
}

func NewGainFood(qty int, foodType FoodType, source FoodSupplier) *GainFoodPower {
	return &GainFoodPower{
		Qty:      qty,
		FoodType: foodType,
		Source:   source,
	}

}

func (p *GainFoodPower) Execute(bird *Bird, player *Player) error {
	if p.FoodType == -1 {
		// TODO: change state to choosing?
	} else {
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
	Food   FoodType
	Qty    int
	Source FoodSupplier
}

func NewCacheFoodPower(food FoodType, qty int, source FoodSupplier) *CacheFoodPower {
	return &CacheFoodPower{
		Qty:    qty,
		Food:   food,
		Source: source,
	}
}

func (p *CacheFoodPower) Execute(bird *Bird, player *Player) error {
	if p.Source != nil {
		if err := p.Source.GetFood(p.Food, p.Qty); err != nil {
			return err
		}
	}
	bird.CacheFood(p.Qty)
	return nil
}

type DrawFromDeckPower struct {
	Qty  int
	Deck Deck
}

func DrawFromDeck(qty int, deck Deck) *DrawFromDeckPower {
	return &DrawFromDeckPower{
		Qty:  qty,
		Deck: deck,
	}
}

func (p *DrawFromDeckPower) Execute(bird *Bird, player *Player) error {
	birds, err := p.Deck.Draw(p.Qty)
	if err != nil {
		return err
	}
	for _, bird := range birds {
		player.GainBird(bird)
	}
	return nil
}

type DrawFromTrayPower struct {
	Qty      int
	BirdTray *BirdTray
}

func DrawFromTray(qty int, tray *BirdTray) *DrawFromTrayPower {
	return &DrawFromTrayPower{
		Qty:      qty,
		BirdTray: tray,
	}
}

func (p *DrawFromTrayPower) Execute(bird *Bird, player *Player) error {
	if p.BirdTray.Len() < p.Qty {
		return ErrNotEnoughCards
	}

	if p.BirdTray.Len() == p.Qty {
		for _, bird := range p.BirdTray.Birds() {
			if _, err := p.BirdTray.Get(bird.ID); err != nil {
				return err
			}
			player.GainBird(bird)
		}
	} else {
		// TODO: choose cards
	}

	return nil
}

type TuckFromDeckPower struct {
	Qty  int
	Deck Deck
}

func TuckFromDeck(qty int, deck Deck) *TuckFromDeckPower {
	return &TuckFromDeckPower{
		Qty:  qty,
		Deck: deck,
	}
}

func (p *TuckFromDeckPower) Execute(bird *Bird, player *Player) error {
	if _, err := p.Deck.Draw(p.Qty); err != nil {
		return err
	}
	bird.TuckCards(p.Qty)
	return nil
}

type TuckFromHandPower struct {
	Qty int
}

func TuckFromHand(qty int) *TuckFromHandPower {
	return &TuckFromHandPower{
		Qty: qty,
	}
}

func (p *TuckFromHandPower) Execute(bird *Bird, player *Player) error {
	hand := player.GetBirdCards()

	if len(hand) < p.Qty {
		return ErrNotEnoughCards
	}

	if len(hand) == p.Qty {
		if err := player.KeepBirds(nil); err != nil {
			return err
		}
		bird.TuckCards(p.Qty)
	} else {
		// TODO: choose birds
	}

	return nil
}

type FishingPower struct {
	Qty  int
	Food FoodType
}

func NewFishingPower(qty int, food FoodType) *FishingPower {
	return &FishingPower{
		Food: food,
		Qty:  qty,
	}
}

func (p *FishingPower) Execute(bird *Bird, player *Player) error {
	random := rand.Intn(FOOD_TYPE_COUNT)
	if FoodType(random) == p.Food {
		bird.CacheFood(p.Qty)
	}
	return nil
}

type HuntingPower struct {
	Deck Deck
}

func NewHuntingPower(deck Deck) *HuntingPower {
	return &HuntingPower{
		Deck: deck,
	}
}

func (p *HuntingPower) Execute(bird *Bird, player *Player) error {
	birds, err := p.Deck.Draw(1)
	if err != nil {
		return err
	}
	for _, drawn := range birds {
		if drawn.Wingspan < bird.HuntingPower {
			bird.TuckCards(1)
		}
	}
	return nil
}

type LayEggsPower struct {
	Qty  int
	Nest NestType
}

func NewLayEggsPower(qty int, nest NestType) *LayEggsPower {
	return &LayEggsPower{
		Qty:  qty,
		Nest: nest,
	}
}

func (p *LayEggsPower) Execute(bird *Bird, player *Player) error {
	if p.Nest != -1 {
		for _, bird := range player.GetBirdCards() {
			if bird.NestType == p.Nest {
				if err := bird.LayEggs(p.Qty); err != nil {
					return err
				}
			}
		}
	} else {
		if len(player.GetBirdCards()) == 1 {
			return player.GetBirdCards()[0].LayEggs(p.Qty)
		}
		// TODO: choose bird to lay egg
	}
	return nil
}
