package lantern

import (
	"math"
	"sync"
)

type defaultPolicy struct {
	sync.RWMutex
	coster  *coster
	tinyLfu *tinyLfu
}

func newDefaultPolicy(maxKeyCount, maxCost uint64) *defaultPolicy {
	ret := &defaultPolicy{}
	ret.coster = newCoster(int64(maxCost))
	ret.tinyLfu = newTinyLFU(maxKeyCount)
	return ret
}

// 返回
// 1. 淘汰的hash,cost
// 2. 代表是否已经存入
// 3. 错误
func (c *defaultPolicy) add(hashed uint64, cost int64) ([]*costerPair, bool, error) {
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

	sample := make([]costerPair, 0, SampleCount)

	freq := c.tinyLfu.estimate(hashed)
	evict := make([]*costerPair, 0)
	for c.coster.remain(cost) < 0 {
		sample = c.coster.fillSample(sample, SampleCount)
		minHash, minCost, minFreq, minIndex := c.minSample(sample)

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
		evict = append(evict, &costerPair{minHash, minCost})
	}

	c.coster.add(hashed, cost)
	return evict, true, nil
}

func (c *defaultPolicy) minSample(pairs []costerPair) (uint64, int64, uint64, int) {
	minHash, minCost, minFreq, minIndex := uint64(math.MaxUint64), int64(math.MaxInt64), uint64(math.MaxUint64), 0
	for i := range pairs {
		if freq := c.tinyLfu.estimate(pairs[i].hash); freq < minFreq {
			minHash = pairs[i].hash
			minFreq = freq
			minCost = pairs[i].cost
			minIndex = i
		}
	}
	return minHash, minCost, minFreq, minIndex
}
