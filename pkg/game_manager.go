package pkg

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	INITIAL_BIRDS = 5
	INITIAL_FOOD  = 5
)

var (
	ErrGameNotFound = errors.New("You're probably not playing any games")
)

type GameManager struct {
	games   *sync.Map
	players *sync.Map
}

func NewGameManager() *GameManager {
	return &GameManager{
		games:   new(sync.Map),
		players: new(sync.Map),
	}
}

func (g *GameManager) Create(socket Socket, sockets []Socket) (*Message, error) {
	game, err := NewGame(sockets, time.Minute)
	if err != nil {
		return nil, err
	}

	for _, socket := range sockets {
		value, _ := game.players.Load(socket)
		player := value.(*Player)

		g.games.Store(socket, game)
		g.players.Store(player.ID, game)
	}

	game.Start(time.Minute)
	return nil, nil
}

func (g *GameManager) ChooseBirds(socket Socket, birds []any) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}

	var ids []BirdID
	if err := ParsePayload(birds, &ids); err != nil {
		return nil, err
	}

	return nil, game.ChooseBirds(socket, ids)
}

func (g *GameManager) DiscardFood(socket Socket, params map[string]any) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}

	var chosenFood map[FoodType]int
	if err := ParsePayload(params, &chosenFood); err != nil {
		return nil, err
	}

	ready, err := game.DiscardFood(socket, chosenFood)
	if err != nil {
		return nil, err
	}

	if ready {
		for _, player := range game.TurnOrder() {
			player.socket.Send(Response{
				Type:    GameStarted,
				Payload: player.ID,
			})
		}

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

func (g *GameManager) PlayerInfo(socket Socket, playerId string) (*Message, error) {
	uuid, err := uuid.Parse(playerId)
	if err != nil {
		return nil, err
	}

	value, ok := g.players.Load(uuid)
	if !ok {
		return nil, ErrGameNotFound
	}

	game := value.(*Game)
	player := game.GetPlayer(uuid)

	if player == nil {
		return nil, ErrPlayerNotFound
	}

	// if this socket points to no games, store it for the game found
	if _, ok := g.games.Load(socket); !ok {
		player.socket = socket
		g.games.Store(socket, game)
		game.players.Store(socket, player)
		game.sockets.Store(player, socket)
	}

	current, err := game.CurrentPlayer()
	if err != nil {
		return nil, err
	}

	socket.Send(Response{
		Type: PlayerInfo,
		Payload: PlayerInfoPayload{
			Board:      player.board,
			Food:       player.GetFood(),
			Birds:      player.birds.Birds(),
			Current:    current.ID,
			Turn:       game.currTurn,
			Round:      game.currRound,
			BirdTray:   game.BirdTray(),
			TurnOrder:  game.TurnOrder(),
			BirdFeeder: game.Birdfeeder(),
			MaxTurns:   MAX_TURNS - game.currRound,
			Duration:   game.turnDuration.Seconds(),
		},
	})

	return nil, nil
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

			game.players.Range(func(socket, value any) bool {
				g.games.Delete(socket.(Socket))
				g.players.Delete(value.(*Player).ID)
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

func (g *GameManager) Disconnect(socket Socket) (*Message, error) {
	game, err := g.GetSocketGame(socket)
	if err != nil {
		return nil, err
	}
	if err := game.Disconnect(socket); err != nil {
		return nil, err
	}
	g.games.Delete(socket)
	return nil, nil
}

func (g *GameManager) GetSocketGame(socket Socket) (*Game, error) {
	value, ok := g.games.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}
	return value.(*Game), nil
}
