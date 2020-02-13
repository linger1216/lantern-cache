package lantern

import "sync"

type ringConsumer interface {
	Push(datas []uint64) bool
}

type ringBuffer struct {
	consumer ringConsumer
	data     []uint64
	maxSize  int
}

func newRingBuffer(consumer ringConsumer, maxSize int) *ringBuffer {
	return &ringBuffer{
		consumer: consumer,
		maxSize:  maxSize,
		data:     make([]uint64, 0, maxSize),
	}
}

func (r *ringBuffer) put(hash uint64) {
	r.data = append(r.data, hash)
	if len(r.data) >= r.maxSize {
		// todo
		// 两者都要清空数据, 但有所区别
		// true: 数据发出去处理, 但处理方是异步处理的, 为了不使用之前的缓冲区, 所以新建了一个新的buf
		// 这里没有使用深拷贝, 应该就是为了速度
		// false: 既然没有处理, 就直接丢弃即可
		if r.consumer.Push(r.data) {
			r.data = make([]uint64, 0, r.maxSize)
		} else {
			r.data = r.data[:0]
		}
	}
}

type ringPoll struct {
	pool *sync.Pool
}

func newRingPool(consumer ringConsumer, maxSize int) *ringPoll {
	return &ringPoll{
		pool: &sync.Pool{
			New: func() interface{} {
				return newRingBuffer(consumer, maxSize)
			},
		},
	}
}

func (r *ringPoll) put(hash uint64) {
	x := r.pool.Get().(*ringBuffer)
	x.put(hash)
	r.pool.Put(x)
}
