package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestQueue(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		queue := pkg.NewQueue(2)
		socket := pkg.NewTestSocket()

		reply, err := queue.Add(socket, nil)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if reply != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		response, err := socket.Receive()
		if err != nil {
			t.Fatalf("Could not parse response: %v", err)
		}
		if response.Type != pkg.WaitForMatch {
			t.Errorf("Expected %v, got %v", pkg.WaitForMatch, response.Type)
		}
	})

	t.Run("cannot enqueue multiple times", func(t *testing.T) {
		queue := pkg.NewQueue(5)
		socket := pkg.NewTestSocket()

		queue.Add(socket, nil)
		_, err := queue.Add(socket, nil)

		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrAlreadyInQueue {
			t.Errorf("Expected error %v, got %v", pkg.ErrAlreadyInQueue, err)
		}
	})

	t.Run("dequeue", func(t *testing.T) {
		queue := pkg.NewQueue(5)
		socket := pkg.NewTestSocket()

		queue.Add(socket, nil)

		if _, err := queue.Remove(socket); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("cannot dequeue socket which is not in queue", func(t *testing.T) {
		queue := pkg.NewQueue(5)

		_, err := queue.Remove(pkg.NewTestSocket())
		if err == nil {
			t.Fatal("Expected error, got nothing")
		}
		if err != pkg.ErrSocketNotQueued {
			t.Errorf("Expected error %v, got %v", pkg.ErrSocketNotQueued, err)
		}
	})

	t.Run("match found", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		queue.Add(p1, nil)
		reply, err := queue.Add(p2, nil)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if reply.Method != "Matchmaker.CreateMatch" {
			t.Errorf("Expected method %v, got %v", "Matchmaker.CreateMatch", reply.Method)
		}
	})

	t.Run("removes players from match found", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		queue.Add(p1, nil)
		queue.Add(p2, nil)

		if _, err := queue.Remove(p1); err == nil {
			t.Error("Expected error trying to remove socket which is not queued")
		}
	})

	t.Run("queue in batch", func(t *testing.T) {
		queue := pkg.NewQueue(2)

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		reply, err := queue.Add(nil, []pkg.Socket{p1, p2})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if reply == nil {
			t.Fatal("Expected reply")
		}
		if reply.Method != "Matchmaker.CreateMatch" {
			t.Errorf("Expected method %v, got %v", "Matchmaker.CreateMatch", reply.Method)
		}
	})
}
