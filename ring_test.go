package lantern

import "testing"

type wrongConsumer struct {
}

func (w *wrongConsumer) Push(datas ...uint64) bool {
	return false
}

type rightConsumer struct {
}

func (r *rightConsumer) Push(datas ...uint64) bool {
	return true
}

func TestRing(t *testing.T) {
	ring := newRingBuffer(&rightConsumer{}, 16)
	for i := uint64(0); i < 20; i++ {
		ring.put(i)
	}
}
