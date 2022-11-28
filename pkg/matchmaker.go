package pkg

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNoPlayers      = errors.New("No players informed")
	ErrMatchNotFound  = errors.New("Match not found")
	ErrPlayerNotFound = errors.New("Player not present in any matches")
)

type Match struct {
	players   *sync.Map
	confirmed *RingBuffer
}

func NewMatch(players []Socket) *Match {
	sockets := new(sync.Map)
	for _, socket := range players {
		sockets.Store(socket, true)
	}

	return &Match{
		players:   sockets,
		confirmed: NewRingBuffer(len(players)),
	}
}

func (m *Match) Ready() bool {
	return m.confirmed.Full()
}

func (m *Match) Confirmed() int {
	return m.confirmed.Len()
}

func (m *Match) Accept(socket Socket) error {
	if _, ok := m.players.Load(socket); !ok {
		return ErrPlayerNotFound
	}

	m.confirmed.Push(socket)
	response := Response{Type: WaitOtherPlayers}

	if _, err := socket.Send(response); err != nil {
		return err
	}

	return nil
}

type Matchmaker struct {
	timeout time.Duration
	matches *sync.Map
	timers  *sync.Map
}

func NewMatchmaker(timeout time.Duration) *Matchmaker {
	return &Matchmaker{
		timeout: timeout,
		matches: new(sync.Map),
		timers:  new(sync.Map),
	}
}

func (m *Matchmaker) Accept(socket Socket) (*Message, error) {
	value, ok := m.matches.Load(socket)
	if !ok {
		return nil, ErrMatchNotFound
	}

	match := value.(*Match)
	if err := match.Accept(socket); err != nil {
		return nil, err
	}

	if match.Ready() {
		m.matches.Delete(socket)

		// Stop and removes match timer
		timer, _ := m.timers.LoadAndDelete(match)
		timer.(*time.Timer).Stop()

		return &Message{
			Method: "Game.Create",
			Params: match.confirmed.Values(),
		}, nil
	}

	return nil, nil
}

func (m *Matchmaker) Decline(socket Socket) (*Message, error) {
	value, ok := m.matches.Load(socket)
	if !ok {
		return nil, ErrMatchNotFound
	}

	match := value.(*Match)
	if err := m.declineMatch(match); err != nil {
		return nil, err
	}

	// Stop and removes match timer
	timer, _ := m.timers.LoadAndDelete(match)
	timer.(*time.Timer).Stop()

	if match.Confirmed() > 0 {
		return &Message{
			Method: "Queue.Add",
			Params: match.confirmed.Values(),
		}, nil
	}

	return nil, nil
}

func (m *Matchmaker) CreateMatch(socket Socket, players []Socket) (*Message, error) {
	if len(players) == 0 {
		return nil, ErrNoPlayers
	}

	match := NewMatch(players)
	for _, player := range players {
		m.matches.Store(player, match)

		player.Send(Response{
			Type:    MatchFound,
			Payload: m.timeout.Seconds(),
		})
	}

	// Decline automatically after timeout
	timer := time.AfterFunc(m.timeout, func() {
		m.declineMatch(match)
	})

	// Store match timer
	m.timers.Store(match, timer)

	return nil, nil
}

func (m *Matchmaker) declineMatch(match *Match) error {
	match.players.Range(func(key, _ any) bool {
		player := key.(Socket)
		m.matches.Delete(player)
		player.Send(Response{Type: MatchDeclined})
		return true
	})

	return nil
}
