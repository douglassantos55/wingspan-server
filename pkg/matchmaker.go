package pkg

import "github.com/google/uuid"

type Match struct {
	ID        uuid.UUID
	players   []*Socket
	confirmed []*Socket
}

func (m *Match) Accept(socket *Socket) bool {
	found := false

	for _, player := range m.players {
		if player == socket {
			found = true
		}
	}

	if !found {
		return false
	}

	m.confirmed = append(m.confirmed, socket)
	if len(m.confirmed) >= len(m.players) {
		return true
	}

	return false
}

type Matchmaker struct {
	matches map[string]*Match
}

func NewMatchmaker(numPlayers int) *Matchmaker {
	return &Matchmaker{
		matches: make(map[string]*Match),
	}
}

func (m *Matchmaker) Accept(socket *Socket, matchId string) string {
	match, ok := m.matches[matchId]
	if ok {
		if match.Accept(socket) {
			delete(m.matches, matchId)
			return "aoeusnth"
		}
	}
	return ""
}

func (m *Matchmaker) CreateMatch(players []*Socket) string {
	match := &Match{
		ID:      uuid.New(),
		players: players,
	}
	m.matches[match.ID.String()] = match
	return match.ID.String()
}
