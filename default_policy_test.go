package lantern

import "testing"

func Test_Add(t *testing.T) {
	p := newDefaultPolicy(10, 10)

	for i := 0; i < 10; i++ {
		p.add(uint64(i), 1)
	}

	p.add(uint64(99), 1)
}
