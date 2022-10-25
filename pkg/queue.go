package pkg

import (
	"container/list"
	"errors"
	"sync"
)

var (
	ErrAlreadyInQueue  = errors.New("Socket already enqueued")
	ErrSocketNotQueued = errors.New("Socket not enqueued")
)

type Queue struct {
	maxPlayers int
	mutex      *sync.Mutex
	players    *list.List
	sockets    map[Socket]*list.Element
}

func NewQueue(maxPlayers int) *Queue {
	return &Queue{
		maxPlayers: maxPlayers,
		mutex:      new(sync.Mutex),
		players:    list.New(),
		sockets:    make(map[Socket]*list.Element),
	}
}

func (q *Queue) Add(socket Socket) (*Message, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.sockets[socket]; ok {
		return nil, ErrAlreadyInQueue
	}

	q.sockets[socket] = q.players.PushBack(socket)

	if _, err := socket.Send(Response{Type: WaitForMatch}); err != nil {
		return nil, err
	}

	if q.players.Len() >= q.maxPlayers {
		players := make([]Socket, 0)

		for i := 0; i < q.maxPlayers; i++ {
			player := q.players.Remove(q.players.Front()).(Socket)
			players = append(players, player)
			delete(q.sockets, player)
		}

		return &Message{
			Method: "Matchmaker.CreateMatch",
			Params: players,
		}, nil
	}

	return nil, nil
}

func (q *Queue) Remove(socket Socket) (*Message, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.sockets[socket]; !ok {
		return nil, ErrSocketNotQueued
	}

	q.players.Remove(q.sockets[socket])
	delete(q.sockets, socket)

	return nil, nil
}
