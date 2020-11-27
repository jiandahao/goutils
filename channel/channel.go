package channel

import (
	"context"
	"sync/atomic"
)

// The Channel Closing Principle
//
// When using Go channel, one of the principles is not to close the channel from the receiving end,
// nor to close the channel with multiple concurrent senders. In other words, if sender (sender) is
// only sender or the last active sender of the channel, then you should close the channel at sender's Goroutine,
// notifying receiver (s) (receiver) There is no value to read. Maintaining this principle will ensure that it
// never occurs. Send a value to an already closed channel or close a channel that has been closed.

// It's OK to leave a Go channel open forever and never close it. When the channel is no longer used, it will be garbage collected.
// Note that it is only necessary to close a channel if the receiver is looking for a close. Closing the channel is
// a control signal on the channel indicating that no more data follows.
// [Design Question: Channel Closing](https://groups.google.com/g/golang-nuts/c/pZwdYRGxCIk/m/qpbHxRRPJdUJ)

// SafeChannel a safe channel that could prevent sending on closed channel
type SafeChannel struct {
	channel  chan interface{}
	isClosed int32 // 0: opened, 1: closed
	ctx      context.Context
	cancle   context.CancelFunc
}

// NewSafeChannel new a channel
func NewSafeChannel(size int) *SafeChannel {
	if size < 0 {
		panic("invlaid size, should be larger than or equals to 0")
	}
	ctx, cancle := context.WithCancel(context.Background())
	return &SafeChannel{
		channel: make(chan interface{}, size),
		ctx:     ctx,
		cancle:  cancle,
	}
}

// Push push value into channel
//
// return false if the channel has been closed
func (sc *SafeChannel) Push(n interface{}) bool {
	if atomic.LoadInt32(&sc.isClosed) == 1 {
		return false
	}

	for {
		if atomic.LoadInt32(&sc.isClosed) == 1 {
			return false
		}
		select {
		case <-sc.ctx.Done():
			return false
		case sc.channel <- n:
			return true
		}
	}
}

// Pop pop value from channel
func (sc *SafeChannel) Pop() (interface{}, bool) {
	n, ok := <-sc.channel
	return n, ok
}

// Close close channel
func (sc *SafeChannel) Close() {
	if atomic.CompareAndSwapInt32(&sc.isClosed, 0, 1) {
		sc.cancle()
		close(sc.channel)
	}
}
