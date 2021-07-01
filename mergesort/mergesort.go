package mergesort

import "sort"

// The Interface type describes the requirements for a type using the routines in this package.
type Interface interface {
	sort.Interface
	Get(i int) interface{}      // get element at index i
	Set(i int, val interface{}) // set element with index i
}

// Sort sorts data.
func Sort(data Interface) {
	doSort(data, 0, data.Len()-1, make([]interface{}, data.Len()))
}

func doSort(data Interface, start, end int, temp []interface{}) {
	if start >= end {
		return
	}

	mid := (start + end) / 2
	doSort(data, start, mid, temp)
	doSort(data, mid+1, end, temp)
	doMerge(data, start, mid, end, temp)
}

func doMerge(data Interface, start, mid, end int, temp []interface{}) {
	//temp := make([]interface{}, end-start+1)
	var i int = start
	var j int = mid + 1
	var k int = 0

	for i <= mid && j <= end && k < len(temp) {
		if data.Less(i, j) {
			temp[k] = data.Get(i)
			i++
		} else {
			temp[k] = data.Get(j)
			j++
		}
		k++
	}

	for i <= mid {
		temp[k] = data.Get(i)
		i++
		k++
	}

	for j <= end {
		temp[k] = data.Get(j)
		j++
		k++
	}

	k = 0
	for i := start; i <= end && k < len(temp); i++ {
		data.Set(i, temp[k])
		k++
	}
}

// Convenience types for common cases. All these cases are copied from
// sort package, and then put Get and Set methods to meet the requirements of the Interface.

// IntSlice attaches the methods of Interface to []int
type IntSlice []int

func (x IntSlice) Len() int           { return len(x) }
func (x IntSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x IntSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// Get get value at index i
func (x IntSlice) Get(i int) interface{} { return x[i] }

// Set set value to index i
func (x IntSlice) Set(i int, val interface{}) { x[i] = val.(int) }

// Sort is a convenience method: x.Sort() calls Sort(x).
func (x IntSlice) Sort() { Sort(x) }

// StringSlice attaches the methods of Interface to []string.
type StringSlice []string

func (x StringSlice) Len() int           { return len(x) }
func (x StringSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x StringSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// Get get value at index i
func (x StringSlice) Get(i int) interface{} { return x[i] }

// Set set value to index i
func (x StringSlice) Set(i int, val interface{}) { x[i] = val.(string) }

// Sort is a convenience method: x.Sort() calls Sort(x).
func (x StringSlice) Sort() { Sort(x) }

// Float64Slice implements Interface for a []float64
// with not-a-number (NaN) values ordered before other values.
type Float64Slice []float64

func (x Float64Slice) Len() int { return len(x) }

// Less reports whether x[i] should be ordered before x[j], as required by the sort Interface.
// Note that floating-point comparison by itself is not a transitive relation: it does not
// report a consistent ordering for not-a-number (NaN) values.
// This implementation of Less places NaN values before any others, by using:
//
//	x[i] < x[j] || (math.IsNaN(x[i]) && !math.IsNaN(x[j]))
//
func (x Float64Slice) Less(i, j int) bool { return x[i] < x[j] || (isNaN(x[i]) && !isNaN(x[j])) }
func (x Float64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// isNaN is a copy of math.IsNaN to avoid a dependency on the math package.
func isNaN(f float64) bool {
	return f != f
}

// Get get value at index i
func (x Float64Slice) Get(i int) interface{} { return x[i] }

// Set set value to index i
func (x Float64Slice) Set(i int, val interface{}) { x[i] = val.(float64) }

// Sort is a convenience method: x.Sort() calls Sort(x).
func (x Float64Slice) Sort() { Sort(x) }
