package pkg

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrEggLimitReached = errors.New("bird's egg limit reached")
)

type Habitat int

const (
	Forest Habitat = iota
	Grassland
	Wetland
)

type Bird struct {
	ID       int
	Name     string
	EggLimit int
	EggCount int
	Habitat  Habitat
}

func (b *Bird) LayEgg() error {
	if b.EggCount >= b.EggLimit {
		return ErrEggLimitReached
	}
	b.EggCount++
	return nil
}

type BirdTray struct {
	birds *sync.Map // Map of IDs to Bird references
	len   int32     // Current number of birds on tray
	size  int32     // Number of slots available
}

func NewBirdTray(size int32) *BirdTray {
	return &BirdTray{
		size:  size,
		birds: new(sync.Map),
	}
}

func (t *BirdTray) Reset(source Deck) error {
	// discards all birds from tray
	atomic.StoreInt32(&t.len, 0)

	t.birds.Range(func(key, value any) bool {
		t.birds.Delete(key)
		return true
	})

	// refills it with new cards from source
	return t.Refill(source)
}

func (t *BirdTray) Refill(source Deck) error {
	// refills empty spaces with new cards from source
	curr := atomic.LoadInt32(&t.len)
	size := atomic.LoadInt32(&t.size)

	emptySlots := size - curr
	cards, err := source.Draw(int(emptySlots))
	if err != nil {
		return err
	}

	atomic.StoreInt32(&t.len, curr+emptySlots)
	for _, card := range cards {
		t.birds.Store(card.ID, card)
	}
	return nil
}

func (t *BirdTray) Get(id int) (*Bird, error) {
	value, loaded := t.birds.LoadAndDelete(id)
	if !loaded {
		return nil, ErrBirdCardNotFound
	}

	curr := atomic.LoadInt32(&t.len)
	atomic.StoreInt32(&t.len, curr-1)

	return value.(*Bird), nil
}

func (t *BirdTray) Birds() []*Bird {
	birds := make([]*Bird, 0, atomic.LoadInt32(&t.len))
	t.birds.Range(func(key, value any) bool {
		birds = append(birds, value.(*Bird))
		return true
	})
	return birds
}

func (t *BirdTray) Len() int {
	return int(atomic.LoadInt32(&t.len))
}
