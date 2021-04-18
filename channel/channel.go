package channel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
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
//
// Ref:
// 1. [How to Gracefully Close Channels? (How to gracefully close the Go channel?)](https://topic.alibabacloud.com/a/how-to-gracefully-close-channels-how-to-gracefully-close-the-go-channel_1_38_30916423.html)
// 2. [(译)如何优雅的关闭Go Channel](https://www.ulovecode.com/2020/07/14/Go/Golang%E8%AF%91%E6%96%87/%E5%A6%82%E4%BD%95%E4%BC%98%E9%9B%85%E5%85%B3%E9%97%ADGo-Channel/)

// Channel describes a channel
type Channel interface {
	Push(n interface{}) bool  // push value into channel.
	Pop() (interface{}, bool) // pop value from channel.
	Close()                   // close channel immediately without waiting.
	CloseAndWait()            // close the channel and wait all element inside channle to be consumed by consumers.
	Count() int64             // return the current number of elements inside channel.
}

// SafeChannel a safe channel that could prevent sending on closed channel
type SafeChannel struct {
	channel  chan interface{}
	isClosed int32 // 0: opened, 1: closed
	ctx      context.Context
	cancle   context.CancelFunc
	counter  int64
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
			atomic.AddInt64(&sc.counter, 1)
			return true
		}
	}
}

// Pop pop value from channel
func (sc *SafeChannel) Pop() (interface{}, bool) {
	n, ok := <-sc.channel
	if ok {
		atomic.AddInt64(&sc.counter, -1)
	}
	return n, ok
}

// Close close channel
func (sc *SafeChannel) Close() {
	if atomic.CompareAndSwapInt32(&sc.isClosed, 0, 1) {
		sc.cancle()
		close(sc.channel)
	}
}

// CloseAndWait close the channel, and wait until all data in the channel have been poped.
func (sc *SafeChannel) CloseAndWait() {
	if !atomic.CompareAndSwapInt32(&sc.isClosed, 0, 1) {
		return
	}
	sc.cancle()
	for atomic.LoadInt64(&sc.counter) > 0 {
		time.Sleep(time.Millisecond * 100)
	}
	close(sc.channel)
}

// Count returns current number of elements inside channel.
func (sc *SafeChannel) Count() int64 {
	return atomic.LoadInt64(&sc.counter)
}

// RecoverableChannel recoverable channel
type RecoverableChannel struct {
	channel   chan interface{}
	closeOnce sync.Once
	counter   int64
	isClosed  int32 // 0 - opened, 1 - closed / closing
}

// NewRevocerableChannel new a recoverable channel
func NewRevocerableChannel(size int) *RecoverableChannel {
	if size < 0 {
		panic("invlaid size, should be larger than or equals to 0")
	}
	return &RecoverableChannel{
		channel: make(chan interface{}, size),
	}
}

// Push push value into channel
//
// return false if the channel has been closed
func (rc *RecoverableChannel) Push(n interface{}) (ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
			atomic.CompareAndSwapInt32(&rc.isClosed, 0, 1)
			return
		}
	}()

	if atomic.LoadInt32(&rc.isClosed) == 1 {
		return false
	}

	rc.channel <- n
	atomic.AddInt64(&rc.counter, 1)
	return
}

// Pop pop value from channel
func (rc *RecoverableChannel) Pop() (interface{}, bool) {
	n, ok := <-rc.channel
	if ok {
		atomic.AddInt64(&rc.counter, -1)
	}
	return n, ok
}

// Close close channel
func (rc *RecoverableChannel) Close() {
	rc.closeOnce.Do(func() {
		close(rc.channel)
	})
}

// CloseAndWait close the channel, and wait until all data in the channel have been poped.
func (rc *RecoverableChannel) CloseAndWait() {
	for !atomic.CompareAndSwapInt32(&rc.isClosed, 0, 1) {
		time.Sleep(100 * time.Millisecond)
	}

	for atomic.LoadInt64(&rc.counter) > 0 {
		time.Sleep(time.Millisecond * 100)
	}

	rc.Close()
}

// Count returns current number of elements inside channel.
func (rc *RecoverableChannel) Count() int64 {
	return atomic.LoadInt64(&rc.counter)
}
