package pkg

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
)

var (
	ErrMatchNotFound  = errors.New("Match not found")
	ErrPlayerNotFound = errors.New("Player not present in any matches")
)

type Match struct {
	ID        uuid.UUID
	players   []io.Writer
	confirmed []io.Writer
}

func (m *Match) Ready() bool {
	return len(m.confirmed) == len(m.players)
}

func (m *Match) Accept(socket io.Writer) error {
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

	data, err := json.Marshal(Response{Type: WaitOtherPlayers})
	if err != nil {
		return err
	}

	if _, err := socket.Write(data); err != nil {
		return err
	}

	return nil
}

type Matchmaker struct {
	matches map[io.Writer]*Match
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		matches: make(map[io.Writer]*Match),
	}
}

func (m *Matchmaker) Accept(socket io.Writer) (*Message, error) {
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

func (m *Matchmaker) Deny(socket io.Writer) (*Message, error) {
	match, ok := m.matches[socket]
	if !ok {
		return nil, ErrMatchNotFound
	}

	data, err := json.Marshal(Response{Type: MatchDeclined})
	if err != nil {
		return nil, err
	}

	for _, player := range match.players {
		if n, err := player.Write(data); n > 0 && err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (m *Matchmaker) CreateMatch(players []io.Writer) (*Message, error) {
	match := &Match{
		ID:      uuid.New(),
		players: players,
	}
	for _, player := range players {
		m.matches[player] = match
	}
	return nil, nil
}
