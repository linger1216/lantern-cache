package lantern_cache

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	// based on 21million dataset, we observed a maximum key length of 77,
	// with minimum length being 6 and average length being 25. We also
	// observed that 99% of keys had length <64 bytes.
	maxKeyLength = 128
	// workloadSize is the size of array storing sequence of keys that we
	// have in our workload. In the benchmark, we iterate over this array b.N
	// number of times in circular fashion starting at a random position.
	workloadSize = 2 << 20
)

var (
	errKeyNotFound  = errors.New("key not found")
	errInvalidValue = errors.New("invalid value")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func blob(char byte, len int) []byte {
	b := make([]byte, len)
	for index := range b {
		b[index] = char
	}
	return b
}

func bytesList(keyLen int) [][]byte {
	keys := make([][]byte, workloadSize)
	for i := 0; i < workloadSize; i++ {
		keys[i] = blob(byte(i)+'a', keyLen)
	}
	return keys
}

type Cache interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
}

//========================================================================
//                              sync.Map
//========================================================================

type SyncMap struct {
	c *sync.Map
}

func (m *SyncMap) Get(key []byte) ([]byte, error) {
	v, ok := m.c.Load(string(key))
	if !ok {
		return nil, errKeyNotFound
	}

	tv, ok := v.([]byte)
	if !ok {
		return nil, errInvalidValue
	}

	return tv, nil
}

func (m *SyncMap) Set(key, value []byte) error {
	// We are not performing any initialization here unlike other caches
	// given that there is no function available to reset the map.
	m.c.Store(string(key), value)
	return nil
}

func newSyncMap() *SyncMap {
	return &SyncMap{new(sync.Map)}
}

//========================================================================
//                               LanternCache
//========================================================================

type LTCache struct {
	c   *LanternCache
	buf []byte
}

func (b *LTCache) Get(key []byte) ([]byte, error) {
	//return b.c.Get(key)
	return b.c.GetWithBuffer(b.buf, key)
}

func (b *LTCache) Set(key, value []byte) error {
	return b.c.Put(key, value)
}

func newLTCache(bucketCount uint32, maxCapacity uint64, allocatorPolicy string) *LTCache {
	cache := NewLanternCache(&Config{
		BucketCount:          bucketCount,
		ChunkAllocatorPolicy: allocatorPolicy,
		MaxCapacity:          maxCapacity,
		InitCapacity:         maxCapacity / 4,
	})
	buf := make([]byte, 0, 2048)
	for i := 0; i < 2*workloadSize; i++ {
		cache.Put([]byte(strconv.Itoa(i)), []byte("data"))
	}
	cache.Reset()
	return &LTCache{cache, buf}
}

//========================================================================
//                         Benchmark Code
//========================================================================

func runCacheBenchmark(b *testing.B, cache Cache, keys [][]byte, vals [][]byte, pctWrites uint64) {
	b.ReportAllocs()
	size := len(keys)
	mask := size - 1
	rc := uint64(0)

	// initialize cache
	for i := 0; i < size; i++ {
		_ = cache.Set(keys[i], vals[i])
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		index := rand.Int() & mask
		mc := atomic.AddUint64(&rc, 1)

		if pctWrites*mc/100 != pctWrites*(mc-1)/100 {
			for pb.Next() {
				_ = cache.Set(keys[index&mask], vals[index&mask])
				index = index + 1
			}
		} else {
			for pb.Next() {
				_, _ = cache.Get(keys[index&mask])
				index = index + 1
			}
		}
	})
}

func BenchmarkCaches(b *testing.B) {

	G := uint64(1024 * 1024 * 1024)
	//M := uint64(1024 * 1024 * 1024)
	K := uint64(1024)
	_ = K
	//
	//bucketCount uint32
	//maxCapacity uint64
	//allocPolicy string

	benchmarks := []struct {
		bucketCount uint32
		maxCapacity uint64
		allocPolicy string
		keyLen      uint64
		valLen      uint64
		pctWrites   uint64
	}{
		{1, G, "heap", 32, 256, 0},
		{512, G, "heap", 32, 256, 0},
		{1024, G, "heap", 32, 256, 0},
		//
		{1, G, "heap", 32, 512, 0},
		{512, G, "heap", 32, 512, 0},
		{1024, G, "heap", 32, 512, 0},
		//
		{1, G, "heap", 32, K, 0},
		{512, G, "heap", 32, K, 0},
		{1024, G, "heap", 32, K, 0},

		{1, G, "heap", 32, 256, 100},
		{512, G, "heap", 32, 256, 100},
		{1024, G, "heap", 32, 256, 100},

		{1, G, "heap", 32, 512, 100},
		{512, G, "heap", 32, 512, 100},
		{1024, G, "heap", 32, 512, 100},

		{1, G, "heap", 32, K, 100},
		{512, G, "heap", 32, K, 100},
		{1024, G, "heap", 32, K, 100},

		{1, G, "heap", 32, 256, 25},
		{512, G, "heap", 32, 256, 25},
		{1024, G, "heap", 32, 256, 25},

		{1, G, "heap", 32, 512, 25},
		{512, G, "heap", 32, 512, 25},
		{1024, G, "heap", 32, 512, 25},

		{1, G, "heap", 32, K, 25},
		{512, G, "heap", 32, K, 25},
		{1024, G, "heap", 32, K, 25},
	}
	for _, bm := range benchmarks {
		var name string
		if bm.pctWrites == 0 {
			name = "[read]"
		} else if bm.pctWrites == 100 {
			name = "[write]"
		} else {
			name = "[mix]"
		}
		name = fmt.Sprintf("%s bucket:%d capacity:%s alloc:%s kenLen:%d valLen:%d", name, bm.bucketCount, humanSize(int64(bm.maxCapacity)), bm.allocPolicy, bm.keyLen, bm.valLen)
		cache := newLTCache(bm.bucketCount, bm.maxCapacity, bm.allocPolicy)
		keys := bytesList(int(bm.keyLen))
		vals := bytesList(int(bm.valLen))
		b.Run(name, func(b *testing.B) {
			runCacheBenchmark(b, cache, keys, vals, bm.pctWrites)
		})
	}
}

/*
	{"SyncMapRead32-32", newSyncMap(), bytesList(32), bytesList(32), 0},
	{"SyncMapRead64-64", newSyncMap(), bytesList(64), bytesList(64), 0},
	{"SyncMapRead64-128", newSyncMap(), bytesList(64), bytesList(128), 0},
	{"SyncMapRead64-256", newSyncMap(), bytesList(64), bytesList(256), 0},
	{"SyncMapRead64-512", newSyncMap(), bytesList(64), bytesList(512), 0},
	{"SyncMapRead64-1024", newSyncMap(), bytesList(64), bytesList(1024), 0},
*/
