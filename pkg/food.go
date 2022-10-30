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
