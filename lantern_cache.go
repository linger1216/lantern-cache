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

func (lc *LanternCache) String() string {
	var mapLen, mapSize, chunkSize uint64
	for i := range lc.buckets {
		ml, ms, cs := lc.buckets[i].stats()
		mapLen += ml
		mapSize += ms
		chunkSize += cs
	}
	return fmt.Sprintf("%s map len:%d map cap:%s chunk cap:%s",
		lc.stats.Raw(), mapLen, humanSize(int64(mapSize)), humanSize(int64(chunkSize)))
}
