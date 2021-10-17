package redislock

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	"gopkg.in/redis.v5"
)

// RedisLock represents distribute lock implemented by redis
type RedisLock struct {
	*redis.Client
	uuid string // uuid is using to generate a unique token in distribution system
	name string // name of resource that you want to deal with
}

// New creates a redis lock
func New(name string, client *redis.Client) *RedisLock {
	u, _ := uuid.NewV1()
	return &RedisLock{
		Client: client,
		uuid:   u.String(),
		name:   name,
	}
}

var lockLuaScript = `
if redis.call('get',KEYS[1]) == ARGV[1] then
	return 1
else
	return redis.call("set",KEYS[1],ARGV[1],"EX",ARGV[2],"NX")
end
`

// Lock locks RedisLock.
//
// If the lock is already in use, the calling goroutine
// blocks until the lock is available.
func (rl *RedisLock) Lock(expiration time.Duration) {
	for {
		if rl.TryLock(expiration) {
			return
		}

		time.Sleep(time.Millisecond * 50)
	}
}

// TryLock tries to acquire the lock, returns true if succeed, or false if
// the lock is already in use by others.
func (rl *RedisLock) TryLock(expiration time.Duration) bool {
	if _, err := rl.Eval(
		lockLuaScript,
		[]string{rl.name},
		rl.calcToken(),
		int(expiration/time.Millisecond),
	).Result(); err != nil {
		return false
	}

	return true
}

// var unlockLuaScript = `
// if redis.call('get',KEYS[1]) == ARGV[1] then
// 	redis.call('del',KEYS[1]) return 1 else return 0
// end
// `

// lua script to do unlock
// 0: succeed
// 1: could not unlock the lock held by another goroutine
// 2: could not unlock a lock that haven't be held
// 3: failed to unlock
var unlockLuaScript = `
if redis.call('EXISTS', KEYS[1]) == 1 then
 	if redis.call('get',KEYS[1]) == ARGV[1] then
	 	redis.call('del',KEYS[1]) return 0 else return 3
	else
		return 1
	end
else
	return 2
end
`

// Unlock unlocks RedisLock.
//
// A RedisLock is associated with a particular goroutine.
// It is not allowed for one goroutine to lock a RedisLock and then
// arrange for another goroutine to unlock it.
func (rl *RedisLock) Unlock() error {
	v, err := rl.Eval(unlockLuaScript, []string{rl.name}, rl.calcToken()).Result()
	fmt.Println("result", v, err)
	return err
}

func (rl *RedisLock) calcToken() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.TrimPrefix(string(buf[:n]), "goroutine ")
	gid := strings.Fields(idField)[0]

	return fmt.Sprintf("%s:%s", rl.uuid, gid)
}