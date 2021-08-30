package main

import (
	"fmt"
	"time"

	"github.com/jiandahao/goutils/distribution/redislock"
	"github.com/jiandahao/goutils/waitgroup"
	"gopkg.in/redis.v5"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	if err := client.Ping().Err(); err != nil {
		fmt.Println(err)
		return
	}

	var resource int
	rlock := redislock.New("lock_name", client)
	rlock.Unlock()
	if true {
		return
	}
	if rlock.TryLock(time.Second * 10) {
		resource = resource + 1
		defer rlock.Unlock()
	}


	wg := waitgroup.Wrapper{}

	wg.Wrap(func() {
		rlock.Unlock()
	})

	wg.Wait()

	for i := 0; i < 10; i++ {
		wg.Wrap(func() {
			if rlock.TryLock(time.Second * 5) {
				resource = resource + 1
			}
		})
	}
	wg.Wait()

	if resource != 1 {
		panic(fmt.Sprintf("error, resource = %v", resource))
	}

	fmt.Println("good")
}
