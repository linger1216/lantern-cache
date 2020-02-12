package lantern

// 将一个key, hash后传入LFU, 进行频率统计
// LFU要记录所有key, 但全部存放太大, 所以用bloom filter来代替
// 光记录key还不够, 还要记录频率, 这里用count-min来代替
type TinyLfu struct {
	countMinSketch *countMinSketch
	bloomFilter    *bloomFilter
	total          uint64
	currentCount   uint64
}

func newTinyLFU(n uint64) *TinyLfu {
	return &TinyLfu{
		countMinSketch: newCountMinSketch(n),
		bloomFilter:    newBloomFilter(float64(n), 0.01),
		total:          n,
		currentCount:   0,
	}
}

func (t *TinyLfu) Put(keyHash uint64) {
	if add := t.bloomFilter.addIfNotHas(keyHash); add {
		t.countMinSketch.increment(keyHash)
		// 这里把逻辑放在这, 因为判断是否has本来就是有误差, 如果要精确的逻辑
		// 不知道合不合适
		t.currentCount++
		if t.currentCount >= t.total {
			t.reset()
		}
	}

	// maybe here
	//t.currentCount++
	//if t.currentCount >= t.total {
	//	t.reset()
	//}
}

//
func (t *TinyLfu) estimate(keyHash uint64) uint64 {
	hit := t.countMinSketch.estimate(keyHash)
	if t.bloomFilter.has(keyHash) {
		hit++
	}
	return uint64(hit)
}

func (t *TinyLfu) reset() {
	t.countMinSketch.reset()
	t.bloomFilter.reset()
	t.currentCount = 0
}
