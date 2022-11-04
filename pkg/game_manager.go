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
	if err := game.ChooseBirds(socket, birds); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *GameManager) DiscardFood(socket Socket, foodType FoodType, qty int) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	ready, err := game.DiscardFood(socket, foodType, qty)

	if err != nil {
		return nil, err
	}

	if ready {
		if err := game.StartRound(); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// TODO: remove qty and use the player's board to determine how many cards he gets
func (g *GameManager) DrawFromDeck(socket Socket, qty int) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	if err := game.DrawFromDeck(socket, qty); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *GameManager) DrawFromTray(socket Socket, ids []int) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	if err := game.DrawFromTray(socket, ids); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *GameManager) GainFood(socket Socket, foodType FoodType) (*Message, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	if err := game.GainFood(socket, foodType); err != nil {
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
	err := game.EndTurn()

	if err != nil {
		if err == ErrGameOver {
			game.Broadcast(Response{Type: GameOver})

			game.players.Range(func(socket, _ any) bool {
				g.games.Delete(socket.(Socket))
				return true
			})
		}
		if err == ErrRoundEnded {
			game.Broadcast(Response{Type: RoundEnded})
		}
		return nil, nil
	}

	return nil, err
}
