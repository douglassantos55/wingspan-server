package pkg

import (
	"errors"
	"sync"
)

const MAX_ROW_COLUMNS = 5

var (
	ErrRowIsFull       = errors.New("row is full")
	ErrHabitatNotFound = errors.New("habitat not found")
)

type Row struct {
	columns *RingBuffer
}

func NewRow(size int) *Row {
	return &Row{
		columns: NewRingBuffer(size),
	}
}

func (r *Row) IsFull() bool {
	return r.columns.Full()
}

func (r *Row) TotalEggs() int {
	total := 0
	for _, bird := range r.GetBirds() {
		total += bird.EggCount
	}
	return total
}

func (r *Row) GetBirds() []*Bird {
	birds := make([]*Bird, 0)
	for _, curr := range r.columns.Values() {
		bird, ok := curr.(*Bird)
		if ok {
			birds = append(birds, bird)
		}
	}
	return birds
}

func (r *Row) FindBird(id BirdID) *Bird {
	for _, value := range r.columns.Values() {
		bird, ok := value.(*Bird)
		if ok && bird.ID == id {
			return bird
		}
	}
	return nil
}

func (r *Row) PushBird(bird *Bird) error {
	if r.IsFull() {
		return ErrRowIsFull
	}
	r.columns.Push(bird)
	return nil
}

func (r *Row) Exposed() int {
	return r.columns.Len()
}

type Board struct {
	rows *sync.Map
}

func NewBoard() *Board {
	rows := new(sync.Map)

	rows.Store(Forest, NewRow(MAX_ROW_COLUMNS))
	rows.Store(Grassland, NewRow(MAX_ROW_COLUMNS))
	rows.Store(Wetland, NewRow(MAX_ROW_COLUMNS))

	return &Board{rows: rows}
}

// places the bird on the leftmost exposed
// slot of the bird's habitat
func (b *Board) PlayBird(bird *Bird) error {
	value, ok := b.rows.Load(bird.Habitat)
	if !ok {
		return ErrHabitatNotFound
	}
	row := value.(*Row)
	return row.PushBird(bird)
}

// Returns the index of the last exposed column
// for a particular habitat
func (b *Board) Exposed(habitat Habitat) int {
	row, ok := b.rows.Load(habitat)
	if !ok {
		return -1
	}
	return row.(*Row).Exposed()
}

func (b *Board) GetBird(id BirdID) *Bird {
	var bird *Bird

	b.rows.Range(func(_, value any) bool {
		row := value.(*Row)
		bird = row.FindBird(id)
		return bird == nil
	})

	return bird
}

func (b *Board) TotalEggs() int {
	total := 0
	b.rows.Range(func(_, value any) bool {
		total += value.(*Row).TotalEggs()
		return true
	})
	return total
}

func (b *Board) GetBirds() []*Bird {
	birds := make([]*Bird, 0)
	b.rows.Range(func(_, value any) bool {
		birds = append(birds, value.(*Row).GetBirds()...)
		return true
	})
	return birds
}

func (b *Board) GetBirdsWithEggs() map[BirdID]int {
	birds := make(map[BirdID]int)
	for _, bird := range b.GetBirds() {
		if bird.EggCount > 0 {
			birds[bird.ID] = bird.EggCount
		}
	}
	return birds
}

func (b *Board) ActivatePowers(habitat Habitat, player *Player) error {
	value, ok := b.rows.Load(habitat)
	if !ok {
		return ErrHabitatNotFound
	}

	birds := value.(*Row).GetBirds()
	for i := len(birds) - 2; i >= 0; i-- {
		if err := birds[i].CastPower(WhenActivated, player); err != nil {
			return err
		}
	}

	return nil
}
