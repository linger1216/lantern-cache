package lantern

import (
	"fmt"
	"testing"
)

func Test_AddEvit(t *testing.T) {
	p := newDefaultPolicy(10000, 10000)
	for i := 0; i < 10000; i++ {
		pair, save, err := p.add(uint64(randomNumber(1, 10000)), int64(randomNumber(1, 10)))
		if err != nil {
			fmt.Printf("err:%s\n", err.Error())
			return
		}
		for _, v := range pair {
			fmt.Printf("i:%d save:%v evict-hash:%d cost:%d\n", i, save, v.hash, v.cost)
		}
	}
}
