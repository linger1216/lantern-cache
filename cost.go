package lantern

const (
	randomPairCount = 5
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

func (c *coster) randomPair() [randomPairCount]costerPair {
	ret := [randomPairCount]costerPair{}
	i := 0
	for k := range c.m {
		ret[i] = costerPair{k, c.m[k]}
		i++
		if i >= randomPairCount {
			break
		}
	}
	return ret
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
	c.add(hashed, cost-prevCost)
	return true
}

func (c *coster) del(hashed uint64) {
	if cost, ok := c.m[hashed]; ok {
		c.used -= cost
		delete(c.m, hashed)
	}
}
