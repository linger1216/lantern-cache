package lantern

import (
	"sync/atomic"
)

type variable struct {
	doNotUse [9]*uint64
	val      *uint64
}

func newVariable() *variable {
	ret := &variable{}
	for i := 0; i < 9; i++ {
		ret.doNotUse[i] = new(uint64)
	}
	ret.val = new(uint64)
	return ret
}

func (p *variable) add(delta uint64) {
	if p == nil {
		return
	}
	atomic.AddUint64(p.val, delta)
}

func (p *variable) get() uint64 {
	if p == nil {
		return 0
	}
	return atomic.LoadUint64(p.val)
}

type stats struct {
	hit       *variable
	miss      *variable
	keyAdd    *variable
	keyUpdate *variable
	keyEvict  *variable
	costAdd   *variable
	costEvict *variable
	rejects   *variable
}

func newStats() *stats {
	ret := &stats{}
	ret.hit = newVariable()
	ret.miss = newVariable()
	ret.keyAdd = newVariable()
	ret.keyUpdate = newVariable()
	ret.keyEvict = newVariable()
	ret.costAdd = newVariable()
	ret.costEvict = newVariable()
	ret.rejects = newVariable()
	return ret
}
