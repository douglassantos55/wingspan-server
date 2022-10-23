package pkg

import (
	"container/list"
	"encoding/json"
	"errors"
	"io"
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
	sockets    map[io.Writer]*list.Element
}

func NewQueue(maxPlayers int) *Queue {
	return &Queue{
		maxPlayers: maxPlayers,
		mutex:      new(sync.Mutex),
		players:    list.New(),
		sockets:    make(map[io.Writer]*list.Element),
	}
}

func (q *Queue) Add(socket io.Writer) (*Message, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.sockets[socket]; ok {
		return nil, ErrAlreadyInQueue
	}

	q.sockets[socket] = q.players.PushBack(socket)

	data, err := json.Marshal(Response{Type: WaitForMatch})
	if err != nil {
		return nil, err
	}

	if _, err := socket.Write(data); err != nil {
		return nil, err
	}

	if q.players.Len() >= q.maxPlayers {
		players := make([]io.Writer, 0)

		for i := 0; i < q.maxPlayers; i++ {
			player := q.players.Remove(q.players.Front()).(io.ReadWriter)
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

func (q *Queue) Remove(socket io.ReadWriter) (*Message, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.sockets[socket]; !ok {
		return nil, ErrSocketNotQueued
	}

	q.players.Remove(q.sockets[socket])
	delete(q.sockets, socket)

	return nil, nil
}
