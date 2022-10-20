package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestMatchmaker(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(2)

		socket := pkg.NewSocket()
		res := matchmaker.Add(socket)

		if res != "" {
			t.Error("Should not get a response")
		}
	})

	t.Run("match found", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(2)

		p1 := pkg.NewSocket()
		p2 := pkg.NewSocket()

		matchmaker.Add(p1)
		res := matchmaker.Add(p2)

		if res == "" {
			t.Fatal("Expected response, got nothing")
		}
	})

	t.Run("dequeue", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(4)

		socket := pkg.NewSocket()
		matchmaker.Add(socket)

		if !matchmaker.Remove(socket) {
			t.Error("Should remove from queue")
		}
	})

	t.Run("cleanup when match found", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(2)

		p1 := pkg.NewSocket()
		p2 := pkg.NewSocket()

		matchmaker.Add(p1)
		matchmaker.Add(p2)

		if matchmaker.Remove(p1) {
			t.Error("Expected p1 to not be in queue anymore")
		}
		if matchmaker.Remove(p2) {
			t.Error("Expected p2 to not be in queue anymore")
		}
	})
}
