package main

import (
	"fmt"

	"github.com/jiandahao/goutils/mergesort"
)

func main() {
	testdata := []int{9, 4, 2, 6, 8, 1, 3, 7, 5, 10}
	// mergesort.Sort(mergesort.IntSlice(testdata))
	mergesort.IntSlice(testdata).Sort()
	fmt.Println(testdata)

	floatData := []float64{9.0, 4.0, 2.0, 6.0, 8.0, 1, 3, 7, 5, 10}
	mergesort.Float64Slice(floatData).Sort()
	fmt.Println(floatData)

	stringData := []string{"9", "4", "2", "6", "8", "1", "3", "7", "5"}
	mergesort.StringSlice(stringData).Sort()
	fmt.Println(stringData)
}
