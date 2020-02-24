package lantern

import (
	"fmt"
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
	//if n.cleanRound > 1 {
	//	fmt.Printf("w:%d round:%d\n", n.w, n.cleanRound)
	//}
	n.data[n.w] = hashed
	n.w++
	return nil
}

func (n *node) read() (uint64, error) {
	//n.Lock()
	//defer n.Unlock()
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
	//err := e.tail.write(hashed)
	//if err == ErrorNodeFull {
	//	fmt.Printf("no space need alloc\n")
	//	e.Lock()
	//	data, err := e.alloc.getChunk()
	//	if err != nil {
	//		e.Unlock()
	//		return err
	//	}
	//	node := newNode(data)
	//	e.tail.next = node
	//	e.tail = node
	//	e.Unlock()
	//	return e.put(hashed)
	//}
	//return nil
	e.Lock()
	defer e.Unlock()

	err := e.tail.write(hashed)
	if err == ErrorNodeFull {
		fmt.Printf("no space need alloc\n")
		data, err := e.alloc.getChunk()
		if err != nil {
			return err
		}
		node := newNode(data)
		e.tail.next = node
		e.tail = node
		return e.tail.write(hashed)
	}
	return nil
}

// todo
// 要通过store等判断是否过期, 这里先随机判断
// 测试通过后在加上完整逻辑
func (e *expiration) cleanExpire() {
	e.Lock()
	defer e.Unlock()
	p := e.head
	for {
		v, err := p.read()
		if err != nil {
			break
		}
		_ = v
		n := randomNumber(1, 10)
		if n >= 5 {
			p.expire()
			// todo
			// some clean code in store, policy and cost
		}
	}

	// 只有head这一个节点不进行重复利用
	if e.head.next == nil {
		return
	}

	p.cleanRound++
	e.head = e.head.next
	p.next = nil
	e.tail.next = p
	e.tail = e.tail.next
}

func (e *expiration) debug() {
	e.Lock()
	defer e.Unlock()
	p := e.head
	for p != nil {
		if p == e.head {
			fmt.Printf("head:%p r:%d w:%d clean:%d\n", p, p.r, p.w, p.cleanRound)
		} else if p == e.tail {
			fmt.Printf("tail:%p r:%d w:%d clean:%d\n", p, p.r, p.w, p.cleanRound)
		} else {
			fmt.Printf("node:%p r:%d w:%d clean:%d\n", p, p.r, p.w, p.cleanRound)
		}
		p = p.next
	}
}

func (e *expiration) size() int {
	e.Lock()
	defer e.Unlock()
	count := 0
	p := e.head
	for p != nil {
		count++
		p = p.next
	}
	return count
}
