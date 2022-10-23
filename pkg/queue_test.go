package pkg_test

import (
	"encoding/json"
	"io"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestQueue(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		queue := pkg.NewQueue(2)
		socket := pkg.NewTestSocket()

		reply, err := queue.Add(socket)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if reply != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		data, err := io.ReadAll(socket)
		if err != nil {
			t.Fatalf("Failed to read data: %v", err)
		}

		var response pkg.Response
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatalf("Could not parse response: %v", err)
		}
		if response.Type != pkg.WaitForMatch {
			t.Errorf("Expected %v, got %v", pkg.WaitForMatch, response.Type)
		}
	})

	t.Run("cannot enqueue multiple times", func(t *testing.T) {
		queue := pkg.NewQueue(5)
		socket := pkg.NewTestSocket()

		queue.Add(socket)
		_, err := queue.Add(socket)

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

		queue.Add(socket)

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

		queue.Add(p1)
		reply, err := queue.Add(p2)

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

		queue.Add(p1)
		queue.Add(p2)

		if _, err := queue.Remove(p1); err == nil {
			t.Error("Expected error trying to remove socket which is not queued")
		}
	})
}
