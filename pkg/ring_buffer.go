package pkg

import "sync"

type RingBuffer struct {
	mutex  sync.Mutex
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.len++
	r.values[r.tail] = value
	r.tail = r.len % len(r.values)
}

func (r *RingBuffer) Pop() any {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	value := r.values[r.head]
	r.values[r.head] = nil
	r.head = (r.head + 1) % len(r.values)
	return value
}

func (r *RingBuffer) Peek() any {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.values[r.head]
}

func (r *RingBuffer) Full() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.len > 0 && r.head == r.tail
}

func (r *RingBuffer) expand() {
	prev := r.values
	r.values = make([]any, r.len*2)
	copy(r.values, prev)
}