package main

import (
	"time"

	"github.com/jiandahao/goutils/channel"
	"github.com/jiandahao/goutils/waitgroup"
)

func main() {
	wg := waitgroup.Wrapper{}
	//m := sync.Mutex{}
	c := channel.NewSafeChannel(0)
	wg.Wrap(func() {
		for i := 0; i < 100; i++ {
			c.Push(i)
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
		}
	})

	time.Sleep(time.Second * 2)
	c.Close()
	wg.Wait()
}
