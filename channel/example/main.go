package main

import (
	"fmt"
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

func testCloseAndWait(c channel.Channel) {
	doneChan := make(chan struct{})
	go func() {
		for i := 0; i < 20; i++ {
			c.Push(i)
		}
		fmt.Println("done")
		c.CloseAndWait()
		doneChan <- struct{}{}
	}()

	go func() {
		for {
			v, isOpenning := c.Pop()
			if !isOpenning {
				fmt.Println("closed")
				return
			}
			fmt.Printf("%v ", v)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	<-doneChan
}

func main() {
	testChannel(channel.NewSafeChannel(10))
	testChannel(channel.NewRevocerableChannel(10))
	testCloseAndWait(channel.NewSafeChannel(10))
	testCloseAndWait(channel.NewRevocerableChannel(10))
}
