package main

import (
	"fmt"

	"github.com/jiandahao/goutils/container"
)

func main() {
	// min heap, could be used to solve Top-K problem
	minHeap := container.NewPriorityQueue(func(x, y interface{}) bool {
		return x.(int) < y.(int)
	}, container.SetQueueCapacity(10))

	for i := 0; i < 50; i++ {
		minHeap.Push(i)
	}

	fmt.Println("length:", minHeap.Len())

	for i := 0; i < 10; i++ {
		fmt.Println("top:", minHeap.Top(), "pop:", minHeap.Pop())
	}
}
