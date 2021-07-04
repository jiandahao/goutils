package main

import (
	"fmt"

	"github.com/jiandahao/goutils/container/priorityqueue"
)

func main() {
	// min heap, could be used to solve Top-K problem
	minHeap := priorityqueue.New(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	}, priorityqueue.SetCapacity(10))

	for i := 0; i < 50; i++ {
		minHeap.Push(i)
	}

	fmt.Println("length:", minHeap.Len())

	for i := 0; i < 10; i++ {
		fmt.Println("top:", minHeap.Top(), "pop:", minHeap.Pop())
	}
}
