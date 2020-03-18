package lantern

const (
	WrongRate = 0.0001
)

// 将一个key, hash后传入LFU, 进行频率统计
// LFU要记录所有key, 但全部存放太大, 所以用bloom filter来代替
// 光记录key还不够, 还要记录频率, 这里用count-min来代替
type tinyLfu struct {
	countMinSketch *countMinSketch
	bloomFilter    *bloomFilter
}

// n 代表能记录的最大数量
func newTinyLFU(n uint64) *tinyLfu {
	return &tinyLfu{
		countMinSketch: newCountMinSketch(n),
		bloomFilter:    newBloomFilter(float64(n), WrongRate),
	}
}

func (t *tinyLfu) increment(keyHash uint64) {
	if add := t.bloomFilter.addIfNotExist(keyHash); add {
		t.countMinSketch.increment(keyHash)
	}
}

func (t *tinyLfu) estimate(keyHash uint64) uint64 {
	var ret uint64
	if t.bloomFilter.exist(keyHash) {
		ret = uint64(t.countMinSketch.estimate(keyHash))
	}
	return ret
}

func (t *tinyLfu) reset() {
	t.countMinSketch.reset()
	t.bloomFilter.reset()
}
