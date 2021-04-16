package lru

import (
	"sync"
)

// Node node
type Node struct {
	Key   int
	Value interface{}
	Prev  *Node
	Next  *Node
}

// Cache cache
type Cache struct {
	mux      sync.Mutex
	head     *Node
	tail     *Node
	m        map[int]*Node
	capacity int
}

// NewCache new cache
func NewCache(capacity int) Cache {
	head := &Node{}

	tail := &Node{
		Prev: head,
	}

	head.Next = tail
	return Cache{
		capacity: capacity,
		head:     head,
		tail:     tail,
		m:        make(map[int]*Node),
	}
}

// Get get by key
func (lc *Cache) Get(key int) interface{} {
	lc.mux.Lock()
	defer lc.mux.Unlock()

	return lc.get(key)

}

func (lc *Cache) get(key int) interface{} {
	if node, ok := lc.m[key]; ok && node != nil {
		prev := node.Prev
		nxt := node.Next

		prev.Next = nxt
		nxt.Prev = prev

		node.Next = lc.head.Next
		node.Next.Prev = node

		lc.head.Next = node
		node.Prev = lc.head
		return node.Value
	}

	return nil
}

// Put put value into cache
func (lc *Cache) Put(key int, value interface{}) {
	lc.mux.Lock()
	defer lc.mux.Unlock()

	// 如果关键字已经存在，通过get后将重新调整顺序，只需更新数据值即可
	if lc.get(key) != -1 {
		node, _ := lc.m[key]
		node.Value = value
		return
	}

	node := &Node{
		Value: value,
		Key:   key,
	}

	// 最新一个被访问，放在链表头
	node.Next = lc.head.Next
	node.Prev = lc.head
	node.Next.Prev = node

	lc.head.Next = node
	lc.m[key] = node

	if len(lc.m) <= lc.capacity {
		return
	}

	// 超过容量，淘汰尾端的数据
	tail := lc.tail.Prev
	temp := tail.Prev
	temp.Next = lc.tail
	lc.tail.Prev = temp
	delete(lc.m, tail.Key)
}
