package pkg

const (
	MatchFound = "match_found"
)

type Message struct {
	Type    string
	Payload any
}
