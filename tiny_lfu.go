package lantern

// 将一个key, hash后传入LFU, 进行频率统计
// LFU要记录所有key, 但全部存放太大, 所以用bloom filter来代替
// 光记录key还不够, 还要记录频率, 这里用count-min来代替
type tinyLfu struct {
	countMinSketch *countMinSketch
	bloomFilter    *bloomFilter
	total          uint64
	currentCount   uint64
}

func newTinyLFU(n uint64) *tinyLfu {
	return &tinyLfu{
		countMinSketch: newCountMinSketch(n),
		bloomFilter:    newBloomFilter(float64(n), 0.0001),
		total:          n,
		currentCount:   0,
	}
}

func (t *tinyLfu) Put(keyHash uint64) {
	if add := t.bloomFilter.addIfNotHas(keyHash); add {
		t.countMinSketch.increment(keyHash)
		t.currentCount++
		if t.currentCount >= t.total {
			t.reset()
		}
	}
}

// todo
// 原先的设计是get访问也算一次, 比如当前频率是1, 但get的同时, 也会算上一次计数, 变为2
// 但我觉得没必要, 除非以后发现这是个精妙的设计, 但此时并没体会到
func (p *tinyLfu) EstimateOriginal(key uint64) uint64 {
	hits := p.countMinSketch.estimate(key)
	if p.bloomFilter.has(key) {
		hits++
	}
	return uint64(hits)
}

//
func (t *tinyLfu) estimate(keyHash uint64) uint64 {
	var ret uint64
	if t.bloomFilter.has(keyHash) {
		ret = uint64(t.countMinSketch.estimate(keyHash))
	}
	return ret
}

func (t *tinyLfu) reset() {
	t.countMinSketch.reset()
	t.bloomFilter.reset()
	t.currentCount = 0
}
