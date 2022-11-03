package pkg

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrGameOver         = errors.New("Game over")
	ErrRoundEnded       = errors.New("Round ended")
	ErrNoPlayerReady    = errors.New("No player ready")
	ErrFoodNotFound     = errors.New("Food not found")
	ErrNotEnoughFood    = errors.New("Not enough food")
	ErrBirdCardNotFound = errors.New("Bird card not found")
)

const (
	MAX_ROUNDS     = 4
	MAX_TURNS      = 8
	MAX_BIRDS_TRAY = 3
)

type Game struct {
	mutex        sync.Mutex
	currRound    int
	currTurn     int
	firstPlayer  Socket
	deck         Deck
	timer        *time.Timer
	turnDuration time.Duration
	turnOrder    *RingBuffer
	players      *sync.Map
	birdTray     *BirdTray
}

func NewGame(sockets []Socket, turnDuration time.Duration) (*Game, error) {
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

	birdTray := NewBirdTray(MAX_BIRDS_TRAY)
	birdTray.Refill(deck)

	return &Game{
		deck:         deck,
		turnDuration: turnDuration,
		players:      players,
		birdTray:     birdTray,
		turnOrder:    NewRingBuffer(len(sockets)),
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

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.timer = time.AfterFunc(timeout, func() {
		g.Broadcast(Response{Type: GameCanceled})
	})
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

// Discards food and returns whether every player is ready
func (g *Game) DiscardFood(socket Socket, foodType FoodType, qty int) (bool, error) {
	value, ok := g.players.Load(socket)
	if !ok {
		return false, ErrGameNotFound
	}

	player := value.(*Player)
	if err := player.DiscardFood(foodType, qty); err != nil {
		return false, err
	}

	g.turnOrder.Push(socket)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.turnOrder.Full() {
		g.timer.Stop()
	} else {
		socket.Send(Response{
			Type: WaitOtherPlayers,
		})
	}

	return g.turnOrder.Full(), nil
}

func (g *Game) StartRound() error {
	g.mutex.Lock()

	g.currTurn = 0
	g.mutex.Unlock()

	g.firstPlayer = g.turnOrder.Peek().(Socket)
	return g.StartTurn()
}

func (g *Game) StartTurn() error {
	socket, ok := g.turnOrder.Peek().(Socket)
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

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.timer = time.AfterFunc(g.turnDuration, func() {
		g.EndTurn()
	})

	return nil
}

func (g *Game) EndTurn() error {
	g.mutex.Lock()

	g.timer.Stop()
	g.turnOrder.Push(g.turnOrder.Dequeue())

	if g.turnOrder.Peek() == g.firstPlayer {
		g.currTurn++

		if g.currTurn >= (MAX_TURNS - g.currRound) {
			g.mutex.Unlock()
			return g.EndRound()
		}
	}

	g.mutex.Unlock()
	return g.StartTurn()
}

func (g *Game) EndRound() error {
	g.mutex.Lock()

	g.currRound++
	g.turnOrder.Push(g.turnOrder.Dequeue())

	if g.currRound >= MAX_ROUNDS {
		return ErrGameOver
	}

	g.mutex.Unlock()

	g.StartRound()
	g.birdTray.Reset(g.deck)

	return ErrRoundEnded
}

func (g *Game) Broadcast(response Response) {
	g.players.Range(func(key, _ any) bool {
		socket := key.(Socket)
		socket.Send(response)
		return true
	})
}

func (g *Game) BirdTray() []*Bird {
	return g.birdTray.Birds()
}
