package lantern

const (
	SampleCount = 5
)

// 成本计算器
type costerPair struct {
	hash uint64
	cost int64
}

type coster struct {
	m    map[uint64]int64
	max  int64
	used int64
}

func newCoster(max int64) *coster {
	assert(max > 0, "cost max must be bigger than zero")
	return &coster{max: max, used: 0, m: make(map[uint64]int64)}
}

func (c *coster) reset() {
	c.clear()
	c.max = 0
}

func (c *coster) clear() {
	c.m = make(map[uint64]int64)
	c.used = 0
}

func (c *coster) fillSample(in []costerPair) []costerPair {
	if len(in) >= SampleCount {
		return in
	}
	for hash, cost := range c.m {
		in = append(in, costerPair{hash, cost})
		if len(in) >= SampleCount {
			return in
		}
	}
	return in
}

func (c *coster) remain(cost int64) int64 {
	return c.max - c.used - cost
}

func (c *coster) add(hashed uint64, cost int64) bool {
	if c.used+cost > c.max {
		return false
	}
	c.m[hashed] = cost
	c.used += cost
	return true
}

func (c *coster) updateIfHas(hashed uint64, cost int64) bool {
	prevCost, ok := c.m[hashed]
	if !ok {
		return false
	}
	return c.add(hashed, cost-prevCost)
}

func (c *coster) del(hashed uint64) {
	if cost, ok := c.m[hashed]; ok {
		c.used -= cost
		delete(c.m, hashed)
	}
}
