package container

import (
	"container/heap"
	"sync"
)

// PriorityQueue represents a priority queue.
// It is safe for concurrent use by multiple
// goroutines without additional locking or coordination.
type PriorityQueue struct {
	s        *innerSlice
	capacity int // maximum size of queue.
	sync.RWMutex
}

// PriorityQueueOption configs how to initialize a priority queue
type PriorityQueueOption func(pq *PriorityQueue)

// SetQueueCapacity sets the capacity of the queue.
//
// capacity < 0 represents Infinite capacity.
func SetQueueCapacity(capacity int) PriorityQueueOption {
	return func(pq *PriorityQueue) {
		pq.capacity = capacity
	}
}

// NewPriorityQueue new a priority queue.
func NewPriorityQueue(compare compareFunc, opts ...PriorityQueueOption) *PriorityQueue {
	pq := &PriorityQueue{
		s:        newInnerSlice(compare),
		capacity: -1, // infinite capacity by default
	}

	for _, opt := range opts {
		opt(pq)
	}

	return pq
}

// Push pushes elements into queue
func (pq *PriorityQueue) Push(x interface{}) {
	pq.Lock()
	defer pq.Unlock()

	heap.Push(pq.s, x)
	if pq.capacity > 0 && pq.s.Len() > pq.capacity {
		// removes and returns the minimum element (according to compareFunc) from the heap,
		// if the current size of the queue exceeds the maximum capacity.
		heap.Pop(pq.s)
	}
}

// Pop removes and returns the top element.
func (pq *PriorityQueue) Pop() interface{} {
	pq.Lock()
	defer pq.Unlock()

	return heap.Pop(pq.s)
}

// Top accesses the top element (minimum element (according to compareFunc)) from the heap.
func (pq *PriorityQueue) Top() interface{} {
	pq.RLock()
	defer pq.RUnlock()

	if pq.s.Len() > 0 {
		return pq.s.data[0]
	}
	return nil
}

// Len returns the total number of elements.
func (pq *PriorityQueue) Len() int {
	pq.RLock()
	defer pq.RUnlock()

	return pq.s.Len()
}

// compareFunc represents a method using to compare two elements, and it
// should return true if x is considered to go before y.
type compareFunc func(x interface{}, y interface{}) bool

type innerSlice struct {
	data    []interface{}
	compare compareFunc
}

func newInnerSlice(compare compareFunc) *innerSlice {
	return &innerSlice{
		compare: compare,
	}
}

// Len returns the number of elements.
func (s *innerSlice) Len() int {
	return len(s.data)
}

// Less represents sort.Interface.Len, reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
func (s *innerSlice) Less(i int, j int) bool {
	return s.compare(s.data[i], s.data[j])
}

// Swap swaps the elements with indexes i and j.
func (s *innerSlice) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

// Push pushes an elements.
func (s *innerSlice) Push(x interface{}) {
	s.data = append(s.data, x)
}

// Pop removes and returns the element at index s.Len() - 1.
func (s *innerSlice) Pop() interface{} {
	res := s.data[s.Len()-1]
	s.data = s.data[:s.Len()-1]
	return res
}
