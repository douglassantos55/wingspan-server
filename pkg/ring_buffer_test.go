package pkg_test

import (
	"reflect"
	"testing"

	"git.internal.com/wingspan/pkg"
)

func TestRingBuffer(t *testing.T) {
	t.Run("push", func(t *testing.T) {
		buf := pkg.NewRingBuffer(4)

		buf.Push(5)
		buf.Push(7)
		buf.Push(9)

		values := []int{}
		for buf.Peek() != nil {
			values = append(values, buf.Pop().(int))
		}

		expected := []int{5, 7, 9}
		if !reflect.DeepEqual(expected, values) {
			t.Errorf("Expected %v, got %v", expected, values)
		}
	})

	t.Run("loop", func(t *testing.T) {
		buf := pkg.NewRingBuffer(4)

		buf.Push(5)
		buf.Push(6)
		buf.Pop()
		buf.Push(7)
		buf.Push(8)
		buf.Pop()
		buf.Pop()
		buf.Push(9)
		buf.Push(10)

		values := []int{}
		for buf.Peek() != nil {
			values = append(values, buf.Pop().(int))
		}

		expected := []int{8, 9, 10}
		if !reflect.DeepEqual(expected, values) {
			t.Errorf("Expected %v, got %v", expected, values)
		}
	})
}
