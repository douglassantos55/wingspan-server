package pkg

import (
	"encoding/json"
	"errors"
)

var (
	ErrPayloadTypeNotFound = errors.New("Payload type not found")
)

const (
	Error            = "error"
	MatchFound       = "match_found"
	WaitForMatch     = "wait_for_match"
	WaitOtherPlayers = "wait_other_players"
	MatchDeclined    = "match_denied"
	GameStart        = "game_start"
)

type Response struct {
	Type    string
	Payload any
}

type Message struct {
	Method string
	Params any
}

type StartingResources struct {
	Birds []*Bird
	Food  []Food
}

func ParsePayload(payload any, dest any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	switch dest.(type) {
	case *StartingResources:
		return json.Unmarshal(data, dest)
	default:
		return ErrPayloadTypeNotFound
	}
}
