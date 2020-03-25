package lantern

import (
	"fmt"
	"math"
	"sync"
)

type defaultPolicy struct {
	sync.RWMutex
	coster  *coster
	tinyLfu *tinyLfu

	// 异步为了性能
	access chan []uint64
	stop   chan struct{}
}

func newDefaultPolicy(maxKeyCount, maxCost uint64) *defaultPolicy {
	ret := &defaultPolicy{}
	ret.coster = newCoster(int64(maxCost))
	ret.tinyLfu = newTinyLFU(maxKeyCount)
	ret.access = make(chan []uint64, 3)
	ret.stop = make(chan struct{})
	go ret.process()
	return ret
}

func (c *defaultPolicy) close() {
	c.stop <- struct{}{}
	close(c.stop)
	close(c.access)
	fmt.Printf("policy closed\n")
}

func (c *defaultPolicy) process() {
	for {
		select {
		case keys := <-c.access:
			c.Lock()
			//fmt.Printf("[policy] 处理访问记录:%d\n", len(keys))
			c.tinyLfu.bulkIncrement(keys)
			c.Unlock()
		case <-c.stop:
			break
		}
	}
}

func (c *defaultPolicy) pushLfu(keys []uint64) {
	if len(keys) > 0 {
		//fmt.Printf("[policy] 异步发送访问记录:%d\n", len(keys))
		c.access <- keys
	}
}

type policyPair struct {
	hashed uint64
	cost   int64
}

// 返回
// 1. 淘汰的hash,cost
// 2. 代表是否已经存入
// 3. 错误
func (c *defaultPolicy) put(hashed uint64, cost int64) ([]uint64, bool, error) {
	c.Lock()
	defer c.Unlock()

	// 当前成本不能超过设定的最大成本
	if cost >= c.coster.max {
		return nil, false, ErrorCostTooLarge
	}

	// 已经在cache中了, 更新一下cost即可
	if c.coster.updateIfExist(hashed, cost) {
		return nil, false, nil
	}

	// 剩余成本够
	if c.coster.remain(cost) > 0 {
		c.coster.add(hashed, cost)
		return nil, true, nil
	}

	sample := make([]policyPair, 0, SampleCount)

	freq := c.tinyLfu.estimate(hashed)
	evict := make([]uint64, 0)
	for c.coster.remain(cost) < 0 {
		sample = c.coster.fillSample(sample, SampleCount)
		minHash, _, minFreq, minIndex := c.minSample(sample)

		// 随机取了5个成本, 如果随机成本中最少的值都比我们即将要加入的值有价值
		// 那么我们可以拒绝这个值添加
		if freq < minFreq {
			return nil, false, nil
		}

		// 将最没有价值的值淘汰
		c.coster.del(minHash)

		// 假设[3]是成本最小值
		// 1,2,3,4,5  before
		// 1,2,5,4,'' after
		// 空位留待下次补充
		endSamplePos := len(sample) - 1
		if endSamplePos != minIndex {
			sample[minIndex] = sample[endSamplePos]
			sample = sample[:endSamplePos]
		}
		evict = append(evict, minHash)
	}

	c.coster.add(hashed, cost)
	return evict, true, nil
}

func (c *defaultPolicy) minSample(pairs []policyPair) (uint64, int64, uint64, int) {
	minHash, minCost, minFreq, minIndex := uint64(math.MaxUint64), int64(math.MaxInt64), uint64(math.MaxUint64), 0
	for i := range pairs {
		if freq := c.tinyLfu.estimate(pairs[i].hashed); freq < minFreq {
			minHash = pairs[i].hashed
			minFreq = freq
			minCost = pairs[i].cost
			minIndex = i
		}
	}
	return minHash, minCost, minFreq, minIndex
}
