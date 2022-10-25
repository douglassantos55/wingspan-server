package pkg

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMatchNotFound  = errors.New("Match not found")
	ErrPlayerNotFound = errors.New("Player not present in any matches")
)

type Match struct {
	ID        uuid.UUID
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
	matches map[Socket]*Match
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		matches: make(map[Socket]*Match),
	}
}

func (m *Matchmaker) Accept(socket Socket) (*Message, error) {
	match, ok := m.matches[socket]
	if !ok {
		return nil, ErrMatchNotFound
	}

	if err := match.Accept(socket); err != nil {
		return nil, err
	}

	if match.Ready() {
		delete(m.matches, socket)

		return &Message{
			Method: "Game.Create",
			Params: match.confirmed,
		}, nil
	}

	return nil, nil
}

func (m *Matchmaker) Deny(socket Socket) (*Message, error) {
	match, ok := m.matches[socket]
	if !ok {
		return nil, ErrMatchNotFound
	}

	var err error
	for _, player := range match.players {
		delete(m.matches, player)
		if _, e := player.Send(Response{Type: MatchDeclined}); e != nil {
			err = e
		}
	}

	if len(match.confirmed) > 0 {
		return &Message{
			Method: "Queue.Add",
			Params: match.confirmed,
		}, nil
	}

	return nil, err
}

func (m *Matchmaker) CreateMatch(players []Socket) (*Message, error) {
	match := &Match{
		ID:      uuid.New(),
		players: players,
	}
	for _, player := range players {
		m.matches[player] = match
	}
	return nil, nil
}
