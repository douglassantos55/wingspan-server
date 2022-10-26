package pkg

import "sync"

type Game struct {
	deck    *BirdDeck
	players *sync.Map
}

func NewGame(sockets []Socket) (*Game, error) {
	players := new(sync.Map)
	deck := NewDeck(MAX_DECK_SIZE)

	for _, socket := range sockets {
		player := NewPlayer(socket)
		player.GainFood(INITIAL_FOOD)

		if err := player.Draw(deck, INITIAL_BIRDS); err != nil {
			return nil, err
		}
		players.Store(socket, player)
	}

	return &Game{
		deck:    deck,
		players: players,
	}, nil
}

func (g *Game) Start() {
	initialBirds := make([]*Bird, 0)
	for i := 0; i < INITIAL_BIRDS; i++ {
		initialBirds = append(initialBirds, &Bird{ID: i})
	}

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
}

func (g *Game) ChooseBirds(socket Socket, birds []int) error {
	_, ok := g.players.Load(socket)
	if !ok {
		return ErrGameNotFound
	}

	_, err := socket.Send(Response{
		Type:    DiscardFood,
		Payload: len(birds),
	})

	return err
}
