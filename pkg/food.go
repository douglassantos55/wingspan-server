package pkg

import (
	"encoding/json"
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

func (f *Birdfeeder) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.List())
}

func (f *Birdfeeder) GetAll(foodType FoodType) (int, error) {
	value, ok := f.food.Load(foodType)
	if !ok {
		return 0, ErrFoodNotFound
	}
	return value.(int), nil
}

func (f *Birdfeeder) GetFood(foodType FoodType, qty int) error {
	value, ok := f.food.Load(foodType)
	if !ok {
		return ErrFoodNotFound
	}

	if value.(int) < qty {
		return ErrNotEnoughFood
	}

	atomic.AddInt32(&f.len, int32(-qty))
	if value.(int)-qty <= 0 {
		f.food.Delete(foodType)
	} else {
		f.food.Store(foodType, value.(int)-qty)
	}

	return nil
}

func (f *Birdfeeder) Refill() {
	size := atomic.LoadInt32(&f.size)
	curr := atomic.LoadInt32(&f.len)

	for i := 0; i < int(size-curr); i++ {
		foodType := FoodType(rand.Intn(FOOD_TYPE_COUNT))
		curr, loaded := f.food.LoadOrStore(foodType, 1)
		if loaded {
			f.food.Store(foodType, 1+curr.(int))
		}
	}

	atomic.StoreInt32(&f.len, size)
}

func (f *Birdfeeder) Len() int {
	return int(atomic.LoadInt32(&f.len))
}

func (f *Birdfeeder) List() map[FoodType]int {
	items := make(map[FoodType]int)
	f.food.Range(func(key, value any) bool {
		items[key.(FoodType)] = value.(int)
		return true
	})
	return items
}
