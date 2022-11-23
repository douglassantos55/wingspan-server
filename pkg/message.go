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
	ChooseCards      = "choose_cards"
	DiscardFood      = "discard_food"
	GameCanceled     = "game_canceled"
	StartTurn        = "start_turn"
	WaitTurn         = "wait_turn"
	RoundEnded       = "round_ended"
	GameOver         = "game_over"
	FoodGained       = "food_gained"
	BirdsDrawn       = "birds_drawn"
	BirdUpdated      = "bird_updated"
	PayBirdCost      = "pay_bird_cost"
	FoodUpdated      = "food_updated"
	BirdPlayed       = "bird_played"
	ChooseFood       = "choose_food"
	ChooseBirds      = "choose_birds"
)

type Response struct {
	Type    string
	Payload any
}

type Message struct {
	Method string
	Params any
}

type ChooseResources struct {
	Birds []*Bird
	Food  map[FoodType]int
}

type AvailableResources struct {
	EggCost int
	BirdID  BirdID
	Birds   map[BirdID]int
	Food    []FoodType
}

type GainFood struct {
	Amount    int
	Available map[FoodType]int
}

func ParsePayload(payload any, dest any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
