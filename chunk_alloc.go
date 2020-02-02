package lantern_cache

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type chunkFactory interface {
	getChunk(size uint32) ([]byte, error)
}

type heapChunkFactory struct{}

func (h heapChunkFactory) getChunk(size uint32) ([]byte, error) {
	return make([]byte, size), nil
}

type mmapChunkFactory struct{}

func (h mmapChunkFactory) getChunk(size uint32) ([]byte, error) {
	return syscall.Mmap(-1, 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
}

type chunkAllocator struct {
	freeChunks     []*[chunkSize]byte
	freeChunksLock sync.Mutex
	factory        chunkFactory
}

func NewChunkAllocator(policy string) *chunkAllocator {
	ret := &chunkAllocator{}
	if len(policy) == 0 {
		policy = "heap"
	}
	policy = strings.ToLower(policy)
	switch policy {
	case "heap":
		ret.factory = &heapChunkFactory{}
	case "mmap":
		ret.factory = &mmapChunkFactory{}
	default:
		panic(fmt.Errorf("can't support factory %s", policy))
	}
	return ret
}

func (c *chunkAllocator) getChunk() ([]byte, error) {
	c.freeChunksLock.Lock()
	if len(c.freeChunks) == 0 {
		allocSize := chunkSize * chunksPerAlloc
		data, err := c.factory.getChunk(uint32(allocSize))
		if err != nil {
			panic(fmt.Errorf("cannot allocate %d bytes via mmap: %s", chunkSize*chunksPerAlloc, err))
		}
		for len(data) > 0 {
			p := (*[chunkSize]byte)(unsafe.Pointer(&data[0]))
			c.freeChunks = append(c.freeChunks, p)
			data = data[chunkSize:]
		}
	}
	n := len(c.freeChunks) - 1
	p := c.freeChunks[n]

	ret := p[:]

	//ret = ret[:0]  // memset

	c.freeChunks[n] = nil
	c.freeChunks = c.freeChunks[:n]
	c.freeChunksLock.Unlock()
	return ret, nil
}

func (c *chunkAllocator) putChunk(chunk []byte) {
	if chunk == nil {
		return
	}
	chunk = chunk[:chunkSize]
	p := (*[chunkSize]byte)(unsafe.Pointer(&chunk[0]))

	c.freeChunksLock.Lock()
	c.freeChunks = append(c.freeChunks, p)
	c.freeChunksLock.Unlock()
}
