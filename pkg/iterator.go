package pkg

type Iterator[T any] interface {
	Next() T
	Prev() T
}

type SliceIterator[T any] struct {
	curr   int
	values []T
}

func NewSliceIterator[T any](values []T) *SliceIterator[T] {
	return &SliceIterator[T]{
		curr:   -1,
		values: values,
	}
}

func (i *SliceIterator[T]) Next() T {
	i.curr++
	value := i.values[i.curr]
	return value
}

func (i *SliceIterator[T]) Prev() T {
	i.curr--
	value := i.values[i.curr]
	return value
}
