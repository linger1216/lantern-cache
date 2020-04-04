package lantern_cache

import (
	"fmt"
	"time"
)

type LanternCache struct {
	buckets    []*bucket
	hash       Hasher
	bucketMask uint64
	stats      *Stats
}

func NewLanternCache(cfg *Config) *LanternCache {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.BucketCount == 0 {
		cfg.BucketCount = 1024
	}

	if len(cfg.ChunkAllocatorPolicy) == 0 {
		cfg.ChunkAllocatorPolicy = "heap"
	}

	if len(cfg.HashPolicy) == 0 {
		cfg.HashPolicy = "fnv"
	}

	if !isPowerOfTwo(cfg.BucketCount) {
		panic(fmt.Errorf("%d must be power of two", cfg.BucketCount))
	}

	if cfg.MaxCapacity == 0 {
		panic(fmt.Errorf("max capacity has to set"))
	}

	if cfg.InitCapacity == 0 {
		cfg.InitCapacity = cfg.MaxCapacity / 4
	}
	return newLanternCache(cfg)
}

func newLanternCache(cfg *Config) *LanternCache {
	ret := &LanternCache{}
	ret.buckets = make([]*bucket, cfg.BucketCount)
	ret.bucketMask = uint64(cfg.BucketCount) - 1
	ret.hash = NewHasher(cfg.HashPolicy)
	ret.stats = &Stats{}

	chunkAlloc := NewChunkAllocator(cfg.ChunkAllocatorPolicy)
	bucketMaxCapacity := (cfg.MaxCapacity + uint64(cfg.BucketCount) - 1) / uint64(cfg.BucketCount)
	bucketInitCapacity := (cfg.InitCapacity + uint64(cfg.BucketCount) - 1) / uint64(cfg.BucketCount)
	if bucketInitCapacity == 0 {
		bucketInitCapacity++
	}

	bc := &bucketConfig{
		maxCapacity:  bucketMaxCapacity,
		initCapacity: bucketInitCapacity,
		chunkAlloc:   chunkAlloc,
		statistics:   ret.stats,
	}
	for i := range ret.buckets {
		ret.buckets[i] = newBucket(bc)
	}

	//ret.logger.Printf("LanternCache init success max capacity:%s bucket count:%d bucket capacity:%s hash:%s alloc:%s verbose:%v",
	//	humanSize(int64(cfg.MaxCapacity)), len(ret.buckets), humanSize(int64(bucketMaxCapacity)), cfg.HashPolicy, cfg.ChunkAllocatorPolicy, ret.verbose)
	return ret
}

func (lc *LanternCache) Put(key, value []byte) error {
	keyHash := lc.hash.Hash(key)
	bucketIndex := keyHash & lc.bucketMask
	bucket := lc.buckets[bucketIndex]
	return bucket.put(keyHash, key, value, 0)
}

func (lc *LanternCache) PutWithExpire(key, value []byte, expire int64) error {
	keyHash := lc.hash.Hash(key)
	bucketIndex := keyHash & lc.bucketMask
	bucket := lc.buckets[bucketIndex]
	return bucket.put(keyHash, key, value, time.Now().Unix()+expire)
}

func (lc *LanternCache) Get(key []byte) ([]byte, error) {
	keyHash := lc.hash.Hash(key)
	bucketIndex := keyHash & lc.bucketMask
	bucket := lc.buckets[bucketIndex]
	v, err := bucket.get(nil, keyHash, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (lc *LanternCache) GetWithBuffer(dst []byte, key []byte) ([]byte, error) {
	keyHash := lc.hash.Hash(key)
	bucketIndex := keyHash & lc.bucketMask
	bucket := lc.buckets[bucketIndex]
	v, err := bucket.get(dst, keyHash, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (lc *LanternCache) Del(key []byte) {
	keyHash := lc.hash.Hash(key)
	bucketIndex := keyHash & lc.bucketMask
	bucket := lc.buckets[bucketIndex]
	bucket.del(keyHash)
}

func (lc *LanternCache) Reset() {
	for i := range lc.buckets {
		lc.buckets[i].reset()
	}
}

func (lc *LanternCache) Stats() *Stats {
	return lc.stats
}

func (lc *LanternCache) Size() uint64 {
	ret := uint64(0)
	for i := range lc.buckets {
		ret += uint64(lc.buckets[i].size())
	}
	return ret
}

func (lc *LanternCache) Scan(count int) ([][]byte, error) {
	ret := make([][]byte, 0, count)
	for i := range lc.buckets {
		if data, err := lc.buckets[i].scan(count); err == nil && len(data) > 0 {
			ret = append(ret, data...)
		}
		if len(ret) >= count {
			break
		}
	}
	return ret, nil
}

func (lc *LanternCache) String() string {
	var mapLen, mapSize, chunkSize, maxChunkSize uint64
	var bucketMinMapLen, bucketMaxMapLen uint64

	for i := range lc.buckets {
		ml, ms, cs, mcs := lc.buckets[i].stats()
		if i == 0 {
			bucketMinMapLen = ml
			bucketMaxMapLen = ml
		}

		if ml < bucketMinMapLen {
			bucketMinMapLen = ml
		}

		if ml > bucketMaxMapLen {
			bucketMaxMapLen = ml
		}

		mapLen += ml
		mapSize += ms
		chunkSize += cs
		maxChunkSize += mcs
	}
	return fmt.Sprintf("%s mapLen:%d mapCap:%s bucketMinMapLen:%d bucketMaxMapLen:%d bucketAvgMapLen:%d chunkCap:%s maxChunkCap:%s",
		lc.stats.Raw(), mapLen, humanSize(int64(mapSize)), bucketMinMapLen, bucketMaxMapLen, mapLen/uint64(len(lc.buckets)), humanSize(int64(chunkSize)), humanSize(int64(maxChunkSize)))
}
