package lantern

import (
	"math"
	"sync"
)

type caffinePolicy struct {
	sync.RWMutex
	coster  *coster
	tinyLfu *tinyLfu
}

func (c *caffinePolicy) add(hashed uint64, cost int64) ([]*costerPair, bool, error) {
	// go 1.14 defer performance is not problem
	c.Lock()
	defer c.Unlock()

	// 当前成本不能超过设定的最大成本
	if cost >= c.coster.max {
		return nil, false, ErrorCostTooLarge
	}

	// 是否已经在cache中了
	if c.coster.updateIfHas(hashed, cost) {
		return nil, false, nil
	}

	// 剩余成本够
	if c.coster.remain(cost) > 0 {
		c.coster.add(hashed, cost)
		return nil, true, nil
	}

	sample := make([]costerPair, 0, SampleCount)

	hit := c.tinyLfu.estimate(hashed)
	evict := make([]*costerPair, 0)
	for c.coster.remain(cost) < 0 {
		sample = c.coster.fillSample(sample)
		minHash, minCost, minHit, minIndex := c.minSample(sample)
		if hit < minHit {
			// 这里key明明不在里面, hit肯定比较低, 如果比较高, 其实应该是conflict比较高
			// conflict比较高, 是不是也能代表价值, 这个说不好, 是我的理论盲区
			// todo
			return nil, false, nil
		}

		// 当前的hash比较有价值
		c.coster.del(minHash)

		// 将sample对应hash删除, 下一轮继续补充(如果成本还不够的话)
		// 因为fill是append添加, 所以删除最后一个元素, 将min和最后一个元素替换
		lastSamplePos := len(sample) - 1
		if lastSamplePos != minIndex {
			c.coster.m[uint64(minIndex)] = c.coster.m[uint64(lastSamplePos)]
			sample = sample[:lastSamplePos]
		}
		evict = append(evict, &costerPair{minHash, minCost})
	}

	c.coster.add(hashed, cost)
	return evict, true, nil
}

func (c *caffinePolicy) minSample(pairs []costerPair) (uint64, int64, uint64, int) {
	minHash, minCost, minHit, minIndex := uint64(math.MaxUint64), int64(math.MaxInt64), uint64(math.MaxUint64), 0
	for i := range pairs {
		if hit := c.tinyLfu.estimate(pairs[i].hash); hit < minHit {
			minHash = pairs[i].hash
			minHit = hit
			minCost = pairs[i].cost
			minIndex = i
		}
	}
	return minHash, minCost, minHit, minIndex
}