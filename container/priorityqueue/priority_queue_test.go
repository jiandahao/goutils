package priorityqueue

import (
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	// min heap
	queue := New(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	})

	for i := 50; i > 0; i-- {
		queue.Push(i)
	}

	if queue.Len() != 50 {
		t.Fatalf("length should be 50, but got %v", queue.Len())
	}

	for i := 1; i < 50; i++ {
		n := queue.Pop()
		if i != n.(int) {
			t.Fatalf("wanted %v but got %v", i, n)
		}
	}
}

func TestPriorityQueue_SizeLimit(t *testing.T) {
	// max heap
	queue := New(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	}, SetCapacity(10))

	for i := 0; i < 50; i++ {
		queue.Push(i)
	}

	if queue.Len() != 10 {
		t.Fatalf("length should be 10 but get %v", queue.Len())
	}

	// only elements [40-49] should be inside max heap
	for i := 40; i < 40+queue.Len(); i++ {
		n := queue.Pop()
		if i != n.(int) {
			t.Fatalf("wanted %v but got %v", i, n)
		}
	}
}

func BenchmarkPriorityQueue_Push_WithoutCapcityLimit(b *testing.B) {
	queue := New(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	})
	for i := 0; i < b.N; i++ {
		queue.Push(i)
	}
}

func BenchmarkPriorityQueue_Push_WithCapcityLimit(b *testing.B) {
	queue := New(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	}, SetCapacity(10))
	for i := 0; i < b.N; i++ {
		queue.Push(i)
	}
}
