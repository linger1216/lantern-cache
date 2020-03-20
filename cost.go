package lantern

const (
	SampleCount = 5
)

type coster struct {
	m    map[uint64]int64
	max  int64
	used int64
}

func newCoster(max int64) *coster {
	return &coster{max: max, used: 0, m: make(map[uint64]int64)}
}

func (c *coster) getSample(count uint) []*entry {
	ret := make([]*entry, 0, count)
	for hash, cost := range c.m {
		ret = append(ret, &entry{key: hash, cost: cost})
		if len(ret) >= int(count) {
			return ret
		}
	}
	return ret
}

func (c *coster) add(hashed uint64, cost int64) bool {
	if c.used+cost > c.max {
		return false
	}
	c.m[hashed] = cost
	c.used += cost
	return true
}

func (c *coster) update(hashed uint64, cost int64) bool {
	var prevCost int64
	if _, ok := c.m[hashed]; ok {
		prevCost = c.m[hashed]
	}
	dist := prevCost - cost
	c.m[hashed] = cost
	c.used += dist
	return true
}

func (c *coster) updateIfExist(hashed uint64, cost int64) bool {
	if _, ok := c.m[hashed]; ok {
		return c.update(hashed, cost)
	}
	return false
}

func (c *coster) fillSample(in []*entry, count uint) []*entry {
	if len(in) >= int(count) {
		return in
	}
	for hash, cost := range c.m {
		in = append(in, &entry{key: hash, cost: cost})
		if len(in) >= int(count) {
			return in
		}
	}
	return in
}

func (c *coster) reset() {
	c.m = make(map[uint64]int64)
	c.used = 0
	c.max = 0
}

func (c *coster) remain(cost int64) int64 {
	return c.max - c.used - cost
}

func (c *coster) del(hashed uint64) {
	if cost, ok := c.m[hashed]; ok {
		c.used -= cost
		delete(c.m, hashed)
	}
}
