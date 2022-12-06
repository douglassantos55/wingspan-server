package pkg

import "sync"

type RingBuffer[T any] struct {
	mutex  sync.Mutex
	values []T
	head   int
	tail   int
	len    int
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		values: make([]T, size),
	}
}

func (r *RingBuffer[T]) Len() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.len
}

func (r *RingBuffer[T]) Values() []T {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.values
}

func (r *RingBuffer[T]) Push(value T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.len++
	r.values[r.tail] = value
	r.tail = r.len % len(r.values)
}

func (r *RingBuffer[T]) Pop() T {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.len > 0 {
		r.len--
		r.tail = r.len % len(r.values)
	}

	value := r.values[r.tail]
	r.values[r.tail] = *new(T)

	return value
}

func (r *RingBuffer[T]) Dequeue() T {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	value := r.values[r.head]
	r.values[r.head] = *new(T)
	r.head = (r.head + 1) % len(r.values)

	return value
}

func (r *RingBuffer[T]) Peek() T {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.values[r.head]
}

func (r *RingBuffer[T]) Full() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.len > 0 && r.head == r.tail
}

func (r *RingBuffer[T]) expand() {
	prev := r.values
	r.values = make([]T, r.len*2)
	copy(r.values, prev)
}
