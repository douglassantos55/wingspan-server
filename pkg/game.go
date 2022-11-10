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
	ErrNotEnoughEggs    = errors.New("Not enough eggs")
	ErrBirdCardNotFound = errors.New("Bird card not found")
)

const (
	MAX_ROUNDS      = 4
	MAX_TURNS       = 8
	MAX_BIRDS_TRAY  = 3
	MAX_FOOD_FEEDER = 5
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
	birdFeeder   *Birdfeeder
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
		birdFeeder:   NewBirdfeeder(MAX_FOOD_FEEDER),
	}, nil
}

func (g *Game) Start(timeout time.Duration) {
	g.players.Range(func(key, value any) bool {
		socket := key.(Socket)
		player := value.(*Player)

		socket.Send(Response{
			Type: ChooseCards,
			Payload: ChooseResources{
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

func (g *Game) ChooseBirds(socket Socket, birdsToKeep []BirdID) error {
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

func (g *Game) DrawFromDeck(socket Socket) error {
	value, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	player := value.(*Player)
	qty := player.GetCardsToDraw()
	drawnBirds, err := g.deck.Draw(qty)

	if err != nil {
		return err
	}

	for _, bird := range drawnBirds {
		player.GainBird(bird)
	}

	g.players.Range(func(key, _ any) bool {
		s := key.(Socket)
		if s == socket {
			s.Send(Response{
				Type:    BirdsDrawn,
				Payload: drawnBirds,
			})
		} else {
			s.Send(Response{
				Type:    BirdsDrawn,
				Payload: len(drawnBirds),
			})
		}
		return true
	})

	return nil
}

func (g *Game) DrawFromTray(socket Socket, birdIds []BirdID) error {
	value, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	player := value.(*Player)
	if len(birdIds) != player.GetCardsToDraw() {
		return ErrUnexpectedValue
	}

	drawnBirds := make([]*Bird, 0, len(birdIds))

	for _, id := range birdIds {
		bird, err := g.birdTray.Get(id)
		if err != nil {
			return err
		}

		// add bird to player's hand
		player.GainBird(bird)
		drawnBirds = append(drawnBirds, bird)
	}

	g.players.Range(func(key, _ any) bool {
		s := key.(Socket)
		if s == socket {
			s.Send(Response{
				Type:    BirdsDrawn,
				Payload: drawnBirds,
			})
		} else {
			s.Send(Response{
				Type:    BirdsDrawn,
				Payload: len(birdIds),
			})
		}
		return true
	})

	return nil
}

func (g *Game) GainFood(socket Socket, foodType FoodType) error {
	value, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	if g.birdFeeder.Len() <= 1 {
		g.birdFeeder.Refill()
	}

	if err := g.birdFeeder.GetFood(foodType); err != nil {
		return err
	}

	player := value.(*Player)
	player.GainFood(foodType, 1)

	g.Broadcast(Response{
		Type:    FoodGained,
		Payload: player.GetFood(),
	})

	return nil
}

func (g *Game) GetEggsToLay(socket Socket) (int, error) {
	player, err := g.validateSocket(socket)
	if err != nil {
		return 0, err
	}
	return player.GetEggsToLay(), nil
}

func (g *Game) LayEggOnBird(socket Socket, birdId BirdID) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	bird, err := player.LayEgg(birdId)
	if err != nil {
		return err
	}

	g.Broadcast(Response{
		Type:    BirdUpdated,
		Payload: bird,
	})

	return nil
}

func (g *Game) PlayBird(socket Socket, birdId BirdID) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}
	return player.PlayBird(birdId)
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

	if err := g.birdTray.Refill(g.deck); err != nil {
		return err
	}

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

func (g *Game) Birdfeeder() map[FoodType]int {
	return g.birdFeeder.List()
}

func (g *Game) validateSocket(socket Socket) (*Player, error) {
	curr, ok := g.turnOrder.Peek().(Socket)
	if !ok {
		return nil, ErrNoPlayerReady
	}

	if socket != curr {
		return nil, ErrPlayerNotFound
	}

	value, ok := g.players.Load(socket)
	if !ok {
		return nil, ErrGameNotFound
	}

	return value.(*Player), nil
}
