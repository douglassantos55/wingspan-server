package pkg

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
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
	RoundStarted     = "round_started"
	GameStarted      = "game_started"
	GameOver         = "game_over"
	FoodGained       = "food_gained"
	BirdsDrawn       = "birds_drawn"
	BirdUpdated      = "bird_updated"
	PayBirdCost      = "pay_bird_cost"
	FoodUpdated      = "food_updated"
	BirdPlayed       = "bird_played"
	ChooseFood       = "choose_food"
	ChooseBirds      = "choose_birds"
	PlayerInfo       = "player_info"
)

type Response struct {
	Type    string
	Payload any
}

type Message struct {
	Method string
	Params any
}

type StartTurnPayload struct {
	Turn     int
	BirdTray *BirdTray
	Duration float64
	TimeLeft float64
}

type WaitTurnPayload struct {
	Turn     int
	Duration float64
	TimeLeft float64
	BirdTray *BirdTray
	Current  uuid.UUID
}

type RoundStartedPayload struct {
	Round     int
	Turns     int
	BirdTray  *BirdTray
	TurnOrder []*Player
}

type ChooseResources struct {
	Birds []*Bird
	Time  float64
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

type PlayerInfoPayload struct {
	// Game generics
	Turn       int
	Round      int
	MaxTurns   int
	Duration   float64
	TimeLeft   float64
	Current    uuid.UUID
	BirdTray   []*Bird
	TurnOrder  []*Player
	BirdFeeder map[FoodType]int

	// Player specifics
	Birds []*Bird
	Board *Board
	Food  map[FoodType]int
}

func ParsePayload(payload any, dest any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
