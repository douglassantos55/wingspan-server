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
	players   []Socket
	confirmed []Socket
}

func (m *Match) Ready() bool {
	return len(m.confirmed) == len(m.players)
}

func (m *Match) Accept(socket Socket) error {
	found := false

	for _, player := range m.players {
		if player == socket {
			found = true
		}
	}

	if !found {
		return ErrPlayerNotFound
	}

	m.confirmed = append(m.confirmed, socket)
	response := Response{Type: WaitOtherPlayers}

	if _, err := socket.Send(response); err != nil {
		return err
	}

	return nil
}

type Matchmaker struct {
	timeout time.Duration
	matches *sync.Map
}

func NewMatchmaker(timeout time.Duration) *Matchmaker {
	return &Matchmaker{
		timeout: timeout,
		matches: new(sync.Map),
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

		return &Message{
			Method: "Game.Create",
			Params: match.confirmed,
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

	if len(match.confirmed) > 0 {
		return &Message{
			Method: "Queue.Add",
			Params: match.confirmed,
		}, nil
	}

	return nil, nil
}

func (m *Matchmaker) CreateMatch(players []Socket) (*Message, error) {
	if len(players) == 0 {
		return nil, ErrNoPlayers
	}

	match := &Match{players: players}
	for _, player := range players {
		m.matches.Store(player, match)
	}

	// Decline automatically after timeout
	go func() {
		<-time.After(m.timeout)
		m.declineMatch(match)
	}()

	return nil, nil
}

func (m *Matchmaker) declineMatch(match *Match) error {
	var err error
	for _, player := range match.players {
		m.matches.Delete(player)
		if _, e := player.Send(Response{Type: MatchDeclined}); e != nil {
			err = e
		}
	}
	return err
}
