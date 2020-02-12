package lantern

import (
	"fmt"
)

const (
	mask               = 1<<6 - 1
	addressBitsPerWord = 6
)

type bitset struct {
	bits []uint64
}

func newBitset(size uint64) *bitset {
	assert(isPowerOfTwo(size), "newBitset size:%d must be power of two", size)
	ret := &bitset{}
	// 为了保证size会多一个
	// 含义: size = (size + 63) / 64
	size = (size + mask) >> addressBitsPerWord
	ret.bits = make([]uint64, size)
	return ret
}

func (b *bitset) set(index uint64) {
	bitsIndex := index >> addressBitsPerWord
	b.bits[bitsIndex] |= 1 << (index & mask)
}

func (b *bitset) has(index uint64) bool {
	return (b.bits[index>>addressBitsPerWord]>>(index&mask))&1 == 1
}

func (b *bitset) reset() {
	for i := range b.bits {
		b.bits[i] = 0
	}
}

func (b *bitset) debug() {
	fmt.Printf(" === bloom dump begin ===\n")
	for i, v := range b.bits {
		fmt.Printf("[%d] -> %b\n", i, v)
	}
	fmt.Printf(" === bloom dump end ===\n")
}
