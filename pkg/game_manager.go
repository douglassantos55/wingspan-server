package pkg

import (
	"errors"
	"sync"
	"time"
)

const (
	INITIAL_BIRDS = 5
	INITIAL_FOOD  = 5
)

var (
	ErrGameNotFound = errors.New("You're probably not playing any games")
)

type GameManager struct {
	games *sync.Map
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: new(sync.Map),
	}
}

func (g *GameManager) Create(players []Socket) (*Message, error) {
	game, err := NewGame(players, time.Minute)
	if err != nil {
		return nil, err
	}

	for _, player := range players {
		g.games.Store(player, game)
	}

	game.Start(time.Minute)
	return nil, nil
}

func (g *GameManager) ChooseBirds(socket Socket, birds []int) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	game.ChooseBirds(socket, birds)

	return nil, nil
}

func (g *GameManager) DiscardFood(socket Socket, foodType FoodType, qty int) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	if err := game.DiscardFood(socket, foodType, qty); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *GameManager) EndTurn(socket Socket) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)

	game.EndTurn()
	game.StartTurn()

	return nil, nil
}
