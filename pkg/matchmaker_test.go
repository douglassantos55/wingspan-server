package pkg_test

import (
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestMatchmaker(t *testing.T) {
	t.Run("accept match", func(t *testing.T) {
		matchmaker := pkg.NewMatchmaker(2)

		p1 := pkg.NewSocket()
		p2 := pkg.NewSocket()

		matchId := matchmaker.CreateMatch([]*pkg.Socket{p1, p2})

		if res := matchmaker.Accept(p1, matchId); res != "" {
			t.Errorf("Expected no response, got %v", res)
		}
		if res := matchmaker.Accept(p2, matchId); res == "" {
			t.Error("Expected game ID, got nothing")
		}
		if res := matchmaker.Accept(p1, matchId); res != "" {
			t.Errorf("Expected no response, got %v", res)
		}
	})
}
