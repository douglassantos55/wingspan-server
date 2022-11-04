package pkg

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

type FoodType int

const FOOD_TYPE_COUNT = 5

const (
	Fruit FoodType = iota
	Seed
	Invertebrate
	Fish
	Rodent
)

type Birdfeeder struct {
	food *sync.Map
	size int32
	len  int32
}

func NewBirdfeeder(size int) *Birdfeeder {
	feeder := &Birdfeeder{
		size: int32(size),
		food: new(sync.Map),
	}
	feeder.Refill()
	return feeder
}

func (f *Birdfeeder) GetFood(foodType FoodType) error {
	value, ok := f.food.Load(foodType)
	if !ok {
		return ErrFoodNotFound
	}

	if value.(int) < 1 {
		return ErrNotEnoughFood
	}

	atomic.AddInt32(&f.len, -1)
	if value.(int)-1 <= 0 {
		f.food.Delete(foodType)
	} else {
		f.food.Store(foodType, 1-value.(int))
	}

	return nil
}

func (f *Birdfeeder) Refill() {
	size := atomic.LoadInt32(&f.size)
	curr := atomic.LoadInt32(&f.len)

	for i := 0; i < int(size-curr); i++ {
		f.food.Store(FoodType(rand.Intn(FOOD_TYPE_COUNT)), 1)
	}
	atomic.StoreInt32(&f.len, size)
}

func (f *Birdfeeder) Len() int {
	return int(atomic.LoadInt32(&f.len))
}
