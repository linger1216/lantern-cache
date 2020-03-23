package lantern

import (
	"sync"
)

type ringBuffer struct {
	access  *defaultPolicy
	data    []uint64
	maxSize uint64
}

func newRingBuffer(access *defaultPolicy, maxSize uint64) *ringBuffer {
	return &ringBuffer{
		access:  access,
		maxSize: maxSize,
		data:    make([]uint64, 0, maxSize),
	}
}

func (r *ringBuffer) put(hash uint64) {
	r.data = append(r.data, hash)
	if uint64(len(r.data)) >= r.maxSize {
		r.access.pushLfu(r.data)
		//r.data = make([]uint64, 0, r.maxSize)
		r.data = r.data[:0]
	}
}

type ringPoll struct {
	pool *sync.Pool
}

func newRingPool(access *defaultPolicy, maxSize uint64) *ringPoll {
	return &ringPoll{
		pool: &sync.Pool{
			New: func() interface{} {
				return newRingBuffer(access, maxSize)
			},
		},
	}
}

func (r *ringPoll) put(hash uint64) {
	x := r.pool.Get().(*ringBuffer)
	x.put(hash)
	r.pool.Put(x)
}
