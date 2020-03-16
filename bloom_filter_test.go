package lantern

import (
	"fmt"
	"testing"
)

func Test_D(t *testing.T) {
	for i := 0; i < 10000; i++ {
		entries, rounds := calcSizeByWrongPositives(float64(i*1000), 0.0001)
		fmt.Printf("numentries:%f 0.01 -> entries:%d rounds:%d\n", float64(i*1000), entries, rounds)
	}
}

func Test_SetGet(t *testing.T) {
	bl := newBloomFilter(100, 0.001)
	bl.add(6)
	ret := bl.has(6)
	ret = bl.has(5)
	fmt.Printf("ret:%v", ret)
}
