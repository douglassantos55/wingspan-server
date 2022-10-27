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

func TestGameManager(t *testing.T) {
	t.Run("creates game", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		manager.Create([]pkg.Socket{p1})

		if _, err := manager.ChooseBirds(p1, []int{4}); err != nil {
			t.Errorf("Expecte no error, got %v", err)
		}

		p2 := pkg.NewTestSocket()
		if _, err := manager.ChooseBirds(p2, []int{1}); err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("starts with cards and food", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.StartingResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Failed parsing payload: %v", err)
		}

		if len(payload.Birds) != pkg.INITIAL_BIRDS {
			t.Errorf("Expected %v birds, got %v", pkg.INITIAL_BIRDS, len(payload.Birds))
		}
		if payload.Food.Len() != pkg.INITIAL_FOOD {
			t.Errorf("Expected %v food, got %v", pkg.INITIAL_FOOD, payload.Food.Len())
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

	t.Run("choose initial birds", func(t *testing.T) {
		manager := pkg.NewGameManager()

		p1 := pkg.NewTestSocket()
		p2 := pkg.NewTestSocket()

		manager.Create([]pkg.Socket{p1, p2})
		response := assertResponse(t, p1, pkg.ChooseCards)

		var payload pkg.StartingResources
		if err := pkg.ParsePayload(response.Payload, &payload); err != nil {
			t.Fatalf("Error parsing payload: %v", err)
		}

		chosenBirds := []int{payload.Birds[0].ID, payload.Birds[4].ID}
		if _, err := manager.ChooseBirds(p1, chosenBirds); err != nil {
			t.Errorf("Error chosing birds: %v", err)
		}

		response = assertResponse(t, p1, pkg.DiscardFood)
		amountToDiscard := int(response.Payload.(float64))
		if amountToDiscard != len(chosenBirds) {
			t.Errorf("Expected %v, got %v", len(chosenBirds), amountToDiscard)
		}
	})
}
