package lantern

import (
	"fmt"
	"testing"
)

func TestNewTinyLFU(t *testing.T) {
	for i := 0; i <= 10; i++ {
		n := uint64(randomNumber(0, 1<<6))
		newTinyLFU(n)
	}
}

func TestNewTinyLFU_PutEstimate(t *testing.T) {
	lfu := newTinyLFU(2048)
	lfu.Put(12)
	freq := lfu.estimate(12)
	orig := lfu.EstimateOriginal(12)
	fmt.Printf("lid freq %d\n", freq)
	fmt.Printf("orig freq %d\n", orig)
}
