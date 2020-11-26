package channel

import (
	"sync"
	"sync/atomic"
)

// SafeChannel a safe channel that could prevent sending on closed channel
type SafeChannel struct {
	channel  chan interface{}
	m        sync.Mutex
	isClosed int32 // 0: opened, 1: closed
}

// NewSafeChannel new a channel
func NewSafeChannel(size int) *SafeChannel {
	if size < 0 {
		panic("invlaid size, should be larger than or equals to 0")
	}
	return &SafeChannel{
		channel:  make(chan interface{}, size),
		isClosed: 0,
	}
}

// Push push value into channel
//
// return false if the channel has been closed
func (sc *SafeChannel) Push(n interface{}) bool {
	sc.m.Lock()
	defer sc.m.Unlock()
	if atomic.LoadInt32(&sc.isClosed) == 0 {
		sc.channel <- n
		return true
	}
	return false
}

// Pop pop value from channel
func (sc *SafeChannel) Pop() (interface{}, bool) {
	n, ok := <-sc.channel
	return n, ok
}

// Close close channel
func (sc *SafeChannel) Close() {
	if atomic.CompareAndSwapInt32(&sc.isClosed, 0, 1) {
		sc.m.Lock()
		close(sc.channel)
		sc.m.Unlock()
	}
}
