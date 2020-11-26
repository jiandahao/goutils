package waitgroup

import (
	"sync"
	"sync/atomic"
)

// Wrapper a wait group wrapper with counter
type Wrapper struct {
	sync.WaitGroup
	counter int64
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released. If the counter goes negative, Add panics.
func (w *Wrapper) Add(delta int) {
	w.WaitGroup.Add(delta)
	atomic.AddInt64(&w.counter, int64(delta))
}

// Done decrements the WaitGroup counter by one.
func (w *Wrapper) Done() {
	if atomic.LoadInt64(&w.counter) > 0 {
		atomic.AddInt64(&w.counter, int64(-1))
	}
	w.WaitGroup.Done()
}

// Count returns current numbers of goroutines
func (w *Wrapper) Count() int {
	return int(atomic.LoadInt64(&w.counter))
}

// Wrap wraps a callback, automatically adding and decreasing delta in the beginning and ending of callback execution respectively.
func (w *Wrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
