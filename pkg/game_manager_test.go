package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func assertResponse(t testing.TB, socket *pkg.TestSocket, expected string) pkg.Response {
	response, err := socket.GetResponse()
	if err != nil {
		t.Fatalf("Failed reading response: %v", err)
	}
	if response.Type != expected {
		t.Errorf("Expected response %v, got %v", expected, response.Type)
	}
	return *response
}

func TestGame(t *testing.T) {
	t.Run("starts with cards and food", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.GameStart)

		var payload pkg.StartingResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Failed parsing payload: %v", err)
		}

		if len(payload.Birds) != pkg.INITIAL_BIRDS {
			t.Errorf("EXpected %v birds, got %v", pkg.INITIAL_BIRDS, len(payload.Birds))
		}
		if len(payload.Food) != pkg.INITIAL_FOOD {
			t.Errorf("EXpected %v food, got %v", pkg.INITIAL_FOOD, len(payload.Food))
		}

		seenBirds := make(map[*pkg.Bird]bool)
		for _, bird := range payload.Birds {
			if bird == nil {
				t.Fatal("Expected bird, got nil")
			}
			if _, ok := seenBirds[bird]; ok {
				t.Fatalf("Duplicated bird %v", bird)
			}
			seenBirds[bird] = true
		}
	})
}
