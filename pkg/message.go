package pkg

const (
	Error            = "error"
	MatchFound       = "match_found"
	WaitForMatch     = "wait_for_match"
	WaitOtherPlayers = "wait_other_players"
	MatchDeclined    = "match_denied"
)

type Response struct {
	Type    string
	Payload any
}

type Message struct {
	Method string
	Params any
}
