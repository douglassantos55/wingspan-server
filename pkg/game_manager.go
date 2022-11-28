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

func (g *GameManager) Create(socket Socket, players []Socket) (*Message, error) {
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

func (g *GameManager) ChooseBirds(socket Socket, birds []BirdID) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.ChooseBirds(socket, birds)
}

func (g *GameManager) DiscardFood(socket Socket, foodType FoodType, qty int) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}

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

func (g *GameManager) DrawCards(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.DrawCards(socket)
}

func (g *GameManager) DrawFromDeck(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.DrawFromDeck(socket)
}

func (g *GameManager) DrawFromTray(socket Socket, ids []BirdID) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.DrawFromTray(socket, ids)
}

func (g *GameManager) GainFood(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.GainFood(socket)
}

func (g *GameManager) ChooseFood(socket Socket, chosen map[FoodType]int) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.ChooseFood(socket, chosen)
}

func (g *GameManager) LayEggs(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.LayEggs(socket)
}

func (g *GameManager) LayEggsOnBirds(socket Socket, chosen map[BirdID]int) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.LayEggsOnBirds(socket, chosen)
}

func (g *GameManager) PlayCard(socket Socket, birdId BirdID) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.PlayBird(socket, birdId)
}

func (g *GameManager) PayBirdCost(socket Socket, birdId BirdID, food []FoodType, eggs map[BirdID]int) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.PayBirdCost(socket, birdId, food, eggs)
}

func (g *GameManager) ActivatePower(socket Socket, birdId BirdID) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	return nil, game.ActivatePower(socket, birdId)
}

func (g *GameManager) EndTurn(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}

	err = game.EndTurn()

	if err != nil {
		if err == ErrGameOver {
			winner, losers := game.GetResult()

			winner.Send(Response{
				Type:    GameOver,
				Payload: "You win",
			})

			for _, loser := range losers {
				loser.Send(Response{
					Type:    GameOver,
					Payload: "You lost",
				})
			}

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

func (g *GameManager) GetSocketGame(socket Socket) (*Game, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}
	return value.(*Game), nil
}
