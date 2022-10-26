package pkg

import (
	"container/list"
	"errors"
)

const MAX_DECK_SIZE = 170

var (
	ErrEmptyDeck       = errors.New("Cannot draw card, deck is empty")
	ErrUnexpectedValue = errors.New("Cannot draw card, unexpected value found")
)

type Deck struct {
	cards *list.List
}

func NewDeck(size int) *Deck {
	list := list.New()
	for i := 0; i < size; i++ {
		list.PushBack(&Bird{ID: i})
	}

	return &Deck{
		cards: list,
	}
}

func (d *Deck) Len() int {
	return d.cards.Len()
}

func (d *Deck) Draw() (*Bird, error) {
	if d.cards.Len() == 0 {
		return nil, ErrEmptyDeck
	}

	value := d.cards.Remove(d.cards.Front())
	card, ok := value.(*Bird)
	if !ok {
		return nil, ErrUnexpectedValue
	}

	return card, nil
}
