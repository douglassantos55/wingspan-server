package pkg

type RingBuffer struct {
	values []any
	head   int
	tail   int
	len    int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		values: make([]any, size),
	}
}

func (r *RingBuffer) Push(value any) {
	r.len++
	r.values[r.tail] = value
	r.tail = r.len % len(r.values)
}

func (r *RingBuffer) Pop() any {
	value := r.values[r.head]
	r.head = (r.head + 1) % len(r.values)
	return value
}

func (r *RingBuffer) Peek() any {
	if r.head == r.tail {
		return nil
	}
	return r.values[r.head]
}

func (r *RingBuffer) Full() bool {
	return r.head == r.tail
}

func (r *RingBuffer) expand() {
	prev := r.values
	r.values = make([]any, r.len*2)
	copy(r.values, prev)
}
