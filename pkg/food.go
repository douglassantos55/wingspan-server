package pkg

type FoodType int

const FOOD_TYPE_COUNT = 5

const (
	Fruit FoodType = iota
	Seed
	Invertebrate
	Fish
	Rodent
)

type Food map[FoodType]int

func (f Food) Increment(foodType FoodType, qty int) {
	if actual, ok := f[foodType]; ok {
		f[foodType] = actual + qty
	} else {
		f[foodType] = qty
	}
}

func (f Food) Decrement(foodType FoodType, qty int) error {
	if actual, ok := f[foodType]; !ok {
		return ErrFoodNotFound
	} else if actual < qty {
		return ErrNotEnoughFood
	}
	f[foodType] -= qty
	if f[foodType] == 0 {
		delete(f, foodType)
	}
	return nil
}

func (f Food) Len() int {
	total := 0
	for _, qty := range f {
		total += qty
	}
	return total
}
