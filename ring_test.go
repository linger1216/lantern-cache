package lantern

import "testing"

func TestRingBuffer(t *testing.T) {
	ring := newRingBuffer(newDefaultPolicy(100, 100), 64)
	for i := uint64(0); i < 1000; i++ {
		ring.put(i)
	}
}

func TestRingPoll(t *testing.T) {
	policy := newDefaultPolicy(100, 100)
	p := newRingPool(policy, 64)
	for i := uint64(0); i < 1000; i++ {
		p.put(i)
	}
	policy.close()
}
