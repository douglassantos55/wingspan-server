package pkg

type FoodType int

const (
	Fruit FoodType = iota
	Seed
	Invertebrate
	Fish
	Rodent
)

type Food struct {
	Type FoodType
}
