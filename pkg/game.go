package pkg

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrNoPlayerReady    = errors.New("No player ready")
	ErrFoodNotFound     = errors.New("Food not found")
	ErrNotEnoughFood    = errors.New("Not enough food")
	ErrBirdCardNotFound = errors.New("Bird card not found")
)

type Game struct {
	deck    Deck
	turns   *RingBuffer
	players *sync.Map
}

func NewGame(sockets []Socket) (*Game, error) {
	if len(sockets) == 0 {
		return nil, ErrNoPlayers
	}

	players := new(sync.Map)
	deck := NewDeck(MAX_DECK_SIZE)

	for _, socket := range sockets {
		player := NewPlayer(socket)

		for i := 0; i < INITIAL_FOOD; i++ {
			foodType := FoodType(rand.Intn(FOOD_TYPE_COUNT))
			player.GainFood(foodType, 1)
		}

		if err := player.Draw(deck, INITIAL_BIRDS); err != nil {
			return nil, err
		}
		players.Store(socket, player)
	}

	return &Game{
		deck:    deck,
		players: players,
		turns:   NewRingBuffer(len(sockets)),
	}, nil
}

func (g *Game) Start(timeout time.Duration) {
	g.players.Range(func(key, value any) bool {
		socket := key.(Socket)
		player := value.(*Player)

		socket.Send(Response{
			Type: ChooseCards,
			Payload: StartingResources{
				Food:  player.GetFood(),
				Birds: player.GetBirdCards(),
			},
		})

		return true
	})

	go func() {
		<-time.After(timeout)
		if !g.turns.Full() {
			g.players.Range(func(key, _ any) bool {
				socket := key.(Socket)
				socket.Send(Response{Type: GameCanceled})
				return true
			})
		}
	}()
}

func (g *Game) ChooseBirds(socket Socket, birdsToKeep []int) error {
	value, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	player := value.(*Player)
	if err := player.KeepBirds(birdsToKeep); err != nil {
		return err
	}

	_, err := socket.Send(Response{
		Type:    DiscardFood,
		Payload: len(birdsToKeep),
	})

	return err
}

func (g *Game) DiscardFood(socket Socket, foodType FoodType, qty int) error {
	value, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	player := value.(*Player)
	if err := player.DiscardFood(foodType, qty); err != nil {
		return err
	}

	g.turns.Push(socket)

	if g.turns.Full() {
		return g.StartTurn()
	} else {
		socket.Send(Response{
			Type: WaitOtherPlayers,
		})
	}
	return nil
}

func (g *Game) StartTurn() error {
	socket, ok := g.turns.Pop().(Socket)
	if !ok {
		return ErrNoPlayerReady
	}

	_, ok = g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	g.players.Range(func(key, _ any) bool {
		s := key.(Socket)
		if s == socket {
			s.Send(Response{Type: StartTurn})
		} else {
			s.Send(Response{Type: WaitTurn})
		}
		return true
	})

	return nil
}
