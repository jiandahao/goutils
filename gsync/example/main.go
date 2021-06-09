package main

import (
	"fmt"

	"github.com/jiandahao/goutils/gsync"
)

func main() {
	// map example usage
	m := gsync.Map{}

	m.Store("name", "mike")
	name, ok := m.Load("name")
	if !ok {
		fmt.Println("not found")
		return
	}
	fmt.Println(name)
	fmt.Println("map length", m.Length())

	// slice example usage
	s := gsync.Slice{}
	for i := 0; i < 10; i++ {
		s.Append(i)
	}

	fmt.Println("slice length", s.Length())

	s.Range(func(index int, value interface{}) bool {
		fmt.Println(index, value)
		return true
	})
}
