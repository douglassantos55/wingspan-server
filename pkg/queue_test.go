package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestQueue(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		socket := pkg.NewSocket()
		res, _ := queue.Add(socket)

		if res != nil {
			t.Error("Should not get a response")
		}
	})

	t.Run("cannot enqueue multiple times", func(t *testing.T) {
		queue := pkg.NewQueue(2)
		socket := pkg.NewSocket()

		queue.Add(socket)
		_, err := queue.Add(socket)

		if err == nil {
			t.Fatal("Should not get a response")
		}
		if err != pkg.ErrAlreadyInQueue {
			t.Errorf("Expected error %v, got %v", pkg.ErrAlreadyInQueue, err)
		}
	})

	t.Run("dequeue", func(t *testing.T) {
		queue := pkg.NewQueue(4)
		socket := pkg.NewSocket()

		queue.Add(socket)
		queue.Remove(socket)

		if _, err := queue.Add(socket); err != nil {
			t.Errorf("Should be able to enqueue since it was removed: %v", err)
		}
	})

	t.Run("cannot dequeue socket which is not in queue", func(t *testing.T) {
		queue := pkg.NewQueue(2)
		_, err := queue.Remove(pkg.NewSocket())

		if err == nil {
			t.Fatal("Should not remove socket which is not queued")
		}
		if err != pkg.ErrSocketNotQueued {
			t.Errorf("Expected error %v, got %v", pkg.ErrSocketNotQueued, err)
		}
	})

	t.Run("match found", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		p1 := pkg.NewSocket()
		p2 := pkg.NewSocket()

		queue.Add(p1)
		res, _ := queue.Add(p2)

		if res == nil {
			t.Fatal("Should have a response")
		}
		if res.Type != pkg.MatchFound {
			t.Errorf("Expected type %v, got %v", pkg.MatchFound, res.Type)
		}

		payload := res.Payload.([]*pkg.Socket)
		if len(payload) != 2 {
			t.Errorf("Expected %v players, got %v", 2, len(payload))
		}
	})

	t.Run("removes players from match found", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		p1 := pkg.NewSocket()
		p2 := pkg.NewSocket()

		queue.Add(p1)
		queue.Add(p2)

		if _, err := queue.Remove(p1); err == nil {
			t.Error("Expected error trying to remove socket already removed")
		}
	})
}
