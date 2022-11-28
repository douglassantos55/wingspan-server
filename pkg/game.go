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
	turnOrder    *RingBuffer[Socket]
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
		turnOrder:    NewRingBuffer[Socket](len(sockets)),
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

func (g *Game) DrawCards(socket Socket) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	return player.SetState(&DrawCardsState{
		Qty:    player.GetCardsToDraw(),
		Source: g.birdTray,
	})
}

func (g *Game) DrawFromDeck(socket Socket) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

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
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	if err := player.Process(birdIds); err != nil {
		return err
	}

	g.players.Range(func(key, _ any) bool {
		s := key.(Socket)
		if s == socket {
			s.Send(Response{
				Type:    BirdsDrawn,
				Payload: player.GetBirdCards(),
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

func (g *Game) ChooseFood(socket Socket, chosenFood map[FoodType]int) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	if err := player.Process(chosenFood); err != nil {
		return err
	}

	g.Broadcast(Response{
		Type: FoodGained,
		Payload: map[string]any{
			"player": player.ID,
			"food":   chosenFood,
		},
	})

	return nil
}

func (g *Game) GainFood(socket Socket) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	if g.birdFeeder.Len() <= 1 {
		g.birdFeeder.Refill()
	}

	return player.SetState(&ChooseFoodState{
		Source: g.birdFeeder,
		Qty:    player.GetFoodToGain(),
	})
}

func (g *Game) LayEggs(socket Socket) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	birdIds := make([]BirdID, 0)
	for _, bird := range player.board.GetBirds() {
		birdIds = append(birdIds, bird.ID)
	}

	return player.SetState(&LayEggsState{
		Qty:   player.GetEggsToLay(),
		Birds: birdIds,
	})
}

func (g *Game) LayEggsOnBirds(socket Socket, chosen map[BirdID]int) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	if err := player.Process(chosen); err != nil {
		return err
	}

	birds := make([]*Bird, 0, len(chosen))
	for id := range chosen {
		birds = append(birds, player.board.GetBird(id))
	}

	g.Broadcast(Response{
		Type:    BirdUpdated,
		Payload: chosen,
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

func (g *Game) PayBirdCost(socket Socket, birdId BirdID, food []FoodType, eggs map[BirdID]int) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}

	if err := player.PayBirdCost(birdId, food, eggs); err != nil {
		return err
	}

	g.Broadcast(Response{
		Type: BirdPlayed,
		Payload: map[string]any{
			"player": player.ID,
			"bird":   player.board.GetBird(birdId),
		},
	})

	return nil
}

func (g *Game) ActivatePower(socket Socket, birdId BirdID) error {
	player, err := g.validateSocket(socket)
	if err != nil {
		return err
	}
	bird := player.board.GetBird(birdId)
	if bird == nil {
		return ErrBirdCardNotFound
	}
	return bird.CastPower(WhenActivated, player)
}

func (g *Game) StartRound() error {
	g.mutex.Lock()
	g.currTurn = 0
	g.firstPlayer = g.turnOrder.Peek()
	g.mutex.Unlock()
	return g.StartTurn()
}

func (g *Game) StartTurn() error {
	socket := g.turnOrder.Peek()
	if socket == nil {
		return ErrNoPlayerReady
	}

	_, ok := g.players.Load(socket)
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

func (g *Game) GetResult() (Socket, []Socket) {
	var winner *Player
	winnerScore := -1
	losers := make([]Socket, 0)

	g.players.Range(func(key, value any) bool {
		player := value.(*Player)
		score := player.TotalScore()

		if score > winnerScore {
			winnerScore = score
			winner = player
		} else if score == winnerScore {
			if player.CountFood() > winner.CountFood() {
				winner = player
			}
		} else {
			losers = append(losers, key.(Socket))
		}
		return true
	})

	return winner.socket, losers
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
	curr := g.turnOrder.Peek()
	if curr == nil {
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
