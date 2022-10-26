package pkg

import (
	"container/list"
	"errors"
)

const MAX_DECK_SIZE = 170

var (
	ErrNotEnoughCards  = errors.New("Not enough cards to draw")
	ErrUnexpectedValue = errors.New("Cannot draw card, unexpected value found")
)

type Deck interface {
	Len() int
	Draw(qty int) ([]*Bird, error)
}

type BirdDeck struct {
	cards *list.List
}

func NewDeck(size int) *BirdDeck {
	list := list.New()
	for i := 0; i < size; i++ {
		list.PushBack(&Bird{ID: i})
	}

	return &BirdDeck{
		cards: list,
	}
}

func (d *BirdDeck) Len() int {
	return d.cards.Len()
}

func (d *BirdDeck) Draw(qty int) ([]*Bird, error) {
	if d.cards.Len() < qty {
		return nil, ErrNotEnoughCards
	}

	cards := make([]*Bird, 0)
	for i := 0; i < qty; i++ {
		value := d.cards.Remove(d.cards.Front())
		card, ok := value.(*Bird)
		if !ok {
			return nil, ErrUnexpectedValue
		}
		cards = append(cards, card)
	}

	return cards, nil
}
