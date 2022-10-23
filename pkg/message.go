package pkg

const (
	Error        = "error"
	MatchFound   = "match_found"
	WaitForMatch = "wait_for_match"
)

type Response struct {
	Type    string
	Payload any
}

type Message struct {
	Method string
	Params any
}
