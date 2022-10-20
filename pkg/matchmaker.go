package pkg

import (
	"container/list"

	"github.com/google/uuid"
)

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
	queue    *list.List
	players  int
	matches  map[string]*Match
	elements map[*Socket]*list.Element
}

func NewMatchmaker(numPlayers int) *Matchmaker {
	return &Matchmaker{
		queue:    list.New(),
		players:  numPlayers,
		matches:  make(map[string]*Match),
		elements: make(map[*Socket]*list.Element),
	}
}

// Adds player to queue and returns matchId if
// enough players are in queue
func (m *Matchmaker) Add(socket *Socket) string {
	m.elements[socket] = m.queue.PushBack(socket)

	if m.queue.Len() == m.players {
		players := make([]*Socket, 0)
		for i := 0; i < m.players; i++ {
			player := m.queue.Remove(m.queue.Front()).(*Socket)
			delete(m.elements, player)
			players = append(players, player)
		}

		matchId := m.createMatch(players)
		for _, player := range players {
			if player != socket {
				go player.Send(matchId)
			}
		}

		return matchId
	}

	return ""
}

// Remove a player from queue and return
// whether it was successful
func (m *Matchmaker) Remove(socket *Socket) bool {
	element, ok := m.elements[socket]
	if ok {
		m.queue.Remove(element)
	}
	return ok
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

func (m *Matchmaker) createMatch(players []*Socket) string {
	match := &Match{
		ID:      uuid.New(),
		players: players,
	}
	m.matches[match.ID.String()] = match
	return match.ID.String()
}
