package lantern

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	require.Equal(t, 8192, len(node.data))
}

func TestNewNodeWrite(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)
}

func TestNewNodeWriteRead(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)

	for {
		_, err := node.read()
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 8192, node.r)
}

func TestNewNodeExpire(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)

	for {
		v, err := node.read()
		if err != nil {
			break
		}
		_ = v
		n := randomNumber(1, 10)
		if n >= 5 {
			node.expire()
		}
	}

	require.Equal(t, node.r, node.w)
	if node.r == len(node.data) {
		return
	}

	for i := node.r; i < len(node.data); i++ {
		v := node.data[i]
		require.Equal(t, uint64(0), v)
	}
}

func TestExpirationPut(t *testing.T) {
	e := newExpiration(AllocPolicy)
	N := 100000
	for i := 0; i < N; i++ {
		err := e.put(uint64(i))
		if err != nil {
			break
		}
	}
}

func TestExpirationCleanExpire(t *testing.T) {
	e := newExpiration(AllocPolicy)
	N := 1000000
	for i := 0; i < N; i++ {
		err := e.put(uint64(i))
		if err != nil {
			break
		}
	}

	end := make(chan struct{})
	round := 1 << 10
	count := 0
	go func() {
		for {
			expectHead := e.head.next
			expectTail := e.head
			e.cleanExpire()
			require.Equal(t, e.head, expectHead)
			require.Equal(t, e.tail, expectTail)
			count++
			if count >= round {
				break
			}
			fmt.Printf("round:%d size:%d\n", count, e.size())
		}
		end <- struct{}{}
	}()
	<-end
	e.debug()
}

func TestExpirationComplicate(t *testing.T) {
	e := newExpiration(AllocPolicy)

	writeChan := make(chan struct{})
	expireChan := make(chan struct{})

	go func() {
		N := 1000
		round := 1 << 20
		count := 0
		for count < round {
			for i := 0; i < N; i++ {
				err := e.put(uint64(randomNumber(1, 100000000)))
				if err != nil {
					break
				}
				time.Sleep(time.Millisecond * 1)
			}
			time.Sleep(time.Millisecond * 100)
			//fmt.Printf("write round:%d size:%d\n", count, e.size())
			count++
		}
		writeChan <- struct{}{}
	}()

	go func() {
		round := 1 << 30
		count := 0
		for count < round {
			e.cleanExpire()
			count++
			time.Sleep(time.Second * 2)
			//fmt.Printf("clean round:%d size:%d\n", count, e.size())
		}
		expireChan <- struct{}{}
	}()

	go func() {
		for {
			//fmt.Printf("size:%d\n", e.size())
			e.debug()
			time.Sleep(time.Second)
		}
	}()

	// wait end
	<-writeChan
	<-expireChan
}
