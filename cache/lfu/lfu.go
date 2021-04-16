package lfu

import "sync"

// Node node
type Node struct {
	Key   string
	Value interface{}
	Freq  int
	Next  *Node
	Prev  *Node
}

// LinkedList linked list
type LinkedList struct {
	head *Node
	tail *Node
}

// NewLinkedList new linked list
func NewLinkedList() *LinkedList {
	head := &Node{}
	tail := &Node{
		Prev: head,
	}
	head.Next = tail
	return &LinkedList{
		head: head,
		tail: tail,
	}
}

// PushFront push node after head
func (ll *LinkedList) PushFront(node *Node) {
	if ll == nil || node == nil {
		panic("invalid node or linked list")
	}

	node.Next = ll.head.Next
	node.Next.Prev = node

	node.Prev = ll.head
	ll.head.Next = node
}

// RemoveNode remove node
func (ll *LinkedList) RemoveNode(node *Node) {
	if node == nil {
		panic("invalid node")
	}

	prev := node.Prev
	next := node.Next
	prev.Next = next
	next.Prev = prev
}

// GetLastNode get last node
func (ll *LinkedList) GetLastNode() *Node {
	if ll.tail.Prev == ll.head {
		return nil
	}

	return ll.tail.Prev
}

// IsEmpty returns true if linked list is empty
func (ll *LinkedList) IsEmpty() bool {
	return ll.tail.Prev == ll.head
}

// Cache lfu cache
type Cache struct {
	nodeMap  map[string]*Node
	freqMap  map[int]*LinkedList
	capacity int
	minFreq  int // current minimum frequency
	mux      sync.Mutex
}

// NewCache new lfu cache instance
func NewCache(capacity int) Cache {
	return Cache{
		nodeMap:  make(map[string]*Node),
		freqMap:  make(map[int]*LinkedList),
		capacity: capacity,
		minFreq:  0,
	}
}

// Get get value by key
func (lc *Cache) Get(key string) interface{} {
	lc.mux.Lock()
	defer lc.mux.Unlock()

	node, ok := lc.nodeMap[key]
	if !ok {
		return nil
	}

	lc.updateByFrequency(node)
	return node.Value
}

// Put put data into cache
func (lc *Cache) Put(key string, value interface{}) {
	lc.mux.Lock()
	defer lc.mux.Unlock()

	if node, ok := lc.nodeMap[key]; ok {
		node.Value = value
		lc.updateByFrequency(node)
		return
	}

	if lc.capacity <= 0 {
		return
	}

	if len(lc.nodeMap) >= lc.capacity {
		lc.removeLeastFreqNode()
	}

	lc.updateByFrequency(&Node{
		Key:   key,
		Value: value,
		Freq:  0,
	})
}

func (lc *Cache) updateByFrequency(node *Node) {
	if _, ok := lc.freqMap[node.Freq]; ok {
		lc.freqMap[node.Freq].RemoveNode(node)
		if lc.freqMap[node.Freq].IsEmpty() {
			delete(lc.freqMap, node.Freq)
		}
	}

	node.Freq = node.Freq + 1
	list, ok := lc.freqMap[node.Freq]
	if !ok {
		list = NewLinkedList()
		lc.freqMap[node.Freq] = list
	}

	lc.nodeMap[node.Key] = node
	list.PushFront(node)

	if node.Freq == 1 {
		lc.minFreq = 1
	} else if _, ok := lc.freqMap[lc.minFreq]; !ok {
		lc.minFreq = lc.minFreq + 1
	}
}

func (lc *Cache) removeLeastFreqNode() {
	list := lc.freqMap[lc.minFreq]
	node := list.GetLastNode()

	list.RemoveNode(node)
	if list.IsEmpty() {
		delete(lc.freqMap, node.Freq)
	}

	delete(lc.nodeMap, node.Key)
}
