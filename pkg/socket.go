package pkg

type Socket struct {
	Outgoing chan any
}

func NewSocket() *Socket {
	return &Socket{
		Outgoing: make(chan any),
	}
}

func (s *Socket) Send(data any) {
	s.Outgoing <- data
}
