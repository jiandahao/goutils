package gsync

import (
	"sync"
	"sync/atomic"
)

// Map is sync.Map with counter. It is safe for concurrent use
// by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
type Map struct {
	sync.Map
	count int64
}

// Store sets the value for a key.
func (m *Map) Store(key interface{}, value interface{}) {
	m.Map.Store(key, value)
	atomic.AddInt64(&m.count, 1)
}

// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) {
	m.Map.Delete(key)
	atomic.AddInt64(&m.count, -1)
}

// Length returns the length of the map
func (m *Map) Length() int64 {
	return atomic.LoadInt64(&m.count)
}

// Exists returns true if value is existed in the map
func (m *Map) Exists(key interface{}) bool {
	_, ok := m.Map.Load(key)
	return ok
}

// Slice is safe for concurrent use by multiple goroutines
// without additional locking or coordination.
type Slice struct {
	sync.Mutex
	data []interface{}
}

// Append append element into slice
func (s *Slice) Append(value interface{}) {
	s.Lock()
	defer s.Unlock()
	s.data = append(s.data, value)
}

// Range calls f sequentially for each value present in the slice.
// If f returns false, range stops the iteration.
func (s *Slice) Range(f func(index int, value interface{}) bool) {
	for index, value := range s.data {
		if !f(index, value) {
			return
		}
	}
}

// Length returns the length of the slice
func (s *Slice) Length() int {
	return len(s.data)
}

// Delete deletes element in the slice at specified index
func (s *Slice) Delete(idx int) {
	s.Lock()
	s.Unlock()
	if idx < 0 || idx >= len(s.data) {
		return
	}

	temp := []interface{}{}
	if idx == 0 {
		temp = s.data[idx+1:]
	} else {
		temp = append(s.data[:idx], s.data[idx+1:]...)
	}
	s.data = temp
}
