package main

import (
	"time"

	"github.com/jiandahao/goutils/channel"
	"github.com/jiandahao/goutils/waitgroup"
)

func testChannel(c channel.Channel) {
	wg := waitgroup.Wrapper{}
	wg.Wrap(func() {
		for i := 0; i < 100; i++ {
			if ok := c.Push(i); !ok {
				return
			}
		}
	})

	wg.Wrap(func() {
		for {
			n, ok := c.Pop()
			if !ok {
				break
			}
			_ = n
			if n.(int)%2 == 0 {
				wg.Wrap(func() {
					c.Push(n.(int) + 1)
				})
			}
			time.Sleep(time.Millisecond * 300)
		}
	})

	time.Sleep(time.Second * 2)
	c.Close()
	wg.Wait()
}

func main() {
	testChannel(channel.NewSafeChannel(10))
	testChannel(channel.NewRevocerableChannel(10))
}
