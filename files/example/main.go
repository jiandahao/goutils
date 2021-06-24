package main

import (
	"fmt"

	"github.com/jiandahao/goutils/files"
)

func main() {
	lines, err := files.ReadLines("./main.go")
	if err != nil {
		fmt.Println(err)
		return
	}

	for line := range lines {
		fmt.Println(line)
	}
}
