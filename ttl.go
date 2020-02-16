package lantern

import (
	"sync"
	"unsafe"
)

const (
	AllocPolicy = "heap"
)

type node struct {
	sync.RWMutex
	data       [chunkSize / 8]uint64
	w          int
	r          int
	cleanRound int //经历了几轮清理操作
	next       *node
}

func newNode(data []byte) *node {
	arr := *(*[chunkSize / 8]uint64)(unsafe.Pointer(&data[0]))
	return &node{data: arr, next: nil, w: 0, r: 0}
}

func (n *node) write(hashed uint64) error {
	//n.Lock()
	//defer n.Unlock()
	if n.w >= len(n.data) {
		return ErrorNodeFull
	}
	n.data[n.w] = hashed
	n.w++
	return nil
}

func (n *node) read() (uint64, error) {
	//n.RLock()
	//defer n.RUnlock()
	if n.r >= n.w {
		return 0, ErrorNodeReadEof
	}
	ret := n.data[n.r]
	n.r++
	return ret, nil
}

// 对当前的r废弃
func (n *node) expire() {
	//n.Lock()
	//defer n.Unlock()
	n.r--
	newVal := n.data[n.w-1]
	n.data[n.r] = newVal
	n.data[n.w-1] = 0
	n.w--
}

type expiration struct {
	sync.RWMutex
	alloc *chunkAllocator
	head  *node
	tail  *node
}

func newExpiration(policy string) *expiration {
	alloc := NewChunkAllocator(policy)
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	assert(node != nil, "new code error")
	return &expiration{alloc: alloc, head: node, tail: node}
}

func (e *expiration) put(hashed uint64) error {
	err := e.tail.write(hashed)
	if err == ErrorNodeFull {
		data, err := e.alloc.getChunk()
		if err != nil {
			return err
		}
		node := newNode(data)
		e.tail.next = node
		e.tail = node
		return e.put(hashed)
	}
	return nil
}

func (e *expiration) cleanExpire() {

}
