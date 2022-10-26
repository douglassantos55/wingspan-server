package pkg

import "sync"

const (
	INITIAL_BIRDS = 5
	INITIAL_FOOD  = 5
)

type GameManager struct {
	games *sync.Map
}

func NewGameManager() *GameManager {
	return &GameManager{}
}

func (g *GameManager) Create(players []Socket) (*Message, error) {
	initialBirds := make([]*Bird, 0)
	for i := 0; i < INITIAL_BIRDS; i++ {
		initialBirds = append(initialBirds, &Bird{ID: i})
	}

	for _, player := range players {
		player.Send(Response{
			Type: GameStart,
			Payload: StartingResources{
				Birds: initialBirds,
				Food:  make([]Food, INITIAL_FOOD),
			},
		})
	}
	return nil, nil
}
