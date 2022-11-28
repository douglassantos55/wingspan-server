package pkg

import (
	"errors"
	"sync"
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
	mutex sync.Mutex
	cards *RingBuffer[*Bird]
}

func NewDeck(size int) *BirdDeck {
	buf := NewRingBuffer[*Bird](size)
	for i := 0; i < size; i++ {
		buf.Push(&Bird{
			ID:       BirdID(i),
			EggLimit: size - i - 1,
		})
	}

	return &BirdDeck{
		cards: buf,
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
		cards = append(cards, d.cards.Pop())
	}

	return cards, nil
}
