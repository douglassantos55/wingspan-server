package pkg

import "container/list"

type Matchmaker struct {
	queue    *list.List
	players  int
	elements map[*Socket]*list.Element
}

func NewMatchmaker(numPlayers int) *Matchmaker {
	return &Matchmaker{
		queue:    list.New(),
		players:  numPlayers,
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

func (m *Matchmaker) createMatch(players []*Socket) string {
	return "aoeuaoeua"
}
