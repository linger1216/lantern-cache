package lantern_cache

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestNewLanternCache(t *testing.T) {
	cache := NewLanternCache(nil)
	_ = cache
}

func TestNewLanternCacheBucketCount(t *testing.T) {
	cache := NewLanternCache(&Config{
		BucketCount:          0,
		MaxCapacity:          1,
		ChunkAllocatorPolicy: "",
		HashPolicy:           "",
	})
	if len(cache.buckets) != 1024 {
		t.Fatal("not equal")
	}
}

func TestLanternCachePutGet(t *testing.T) {
	b := NewLanternCache(nil)
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := []byte(fmt.Sprintf("val%d", i))
		err := b.Put(key, val)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := b.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, val) {
			t.Fatal("not equal")
		}
	}
}

func TestLanternCachePutGetWithBuffer(t *testing.T) {
	buf := make([]byte, 0, 64)
	b := NewLanternCache(nil)
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := []byte(fmt.Sprintf("val%d", i))
		err := b.Put(key, val)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := b.GetWithBuffer(buf, key)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, val) {
			t.Fatal("not equal")
		}
	}
}

func TestLanternCachePutGetExpire(t *testing.T) {
	b := NewLanternCache(nil)
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.PutWithExpire(key1, val1, 1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)
	_, err = b.Get(key1)
	if err != ErrorValueExpire {
		t.Fatal(err)
	}
}

func TestLanternCachePutGetSmall(t *testing.T) {
	b := NewLanternCache(nil)
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.Put(key1, val1)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.Get(key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}
}

func TestLanternCacheBigKeyValue(t *testing.T) {
	b := NewLanternCache(nil)
	key1 := []byte("key1")
	val1 := makeByte(64 * 1024)
	err := b.Put(key1, val1)
	if err != ErrorInvalidEntry {
		t.Fatal(err)
	}
}

func TestLanternCacheDel(t *testing.T) {
	b := NewLanternCache(nil)
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.Put(key1, val1)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.Get(key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}

	b.Del(key1)
	_, err = b.Get(key1)
	if err != ErrorNotFound {
		t.Fatal(err)
	}
}

func TestLanternCacheMMap(t *testing.T) {
	b := NewLanternCache(&Config{
		ChunkAllocatorPolicy: "mmap",
		BucketCount:          32,
		MaxCapacity:          1024 * 1024,
	})
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.Put(key1, val1)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.Get(key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}
}

func TestLanternCacheJumpChunk(t *testing.T) {
	b := NewLanternCache(&Config{
		ChunkAllocatorPolicy: "heap",
		BucketCount:          2,
		MaxCapacity:          1024 * 1024,
	})
	val := makeByte(1024)
	for i := 0; i < 1024; i++ {
		err := b.Put([]byte(fmt.Sprintf("key%d", i)), val)
		if err != nil {
			t.Fatal(err)
		}
	}
	if b.buckets[0].chunks[1] == nil {
		t.Fatal("no chunk")
	}

	if b.buckets[0].loop == 0 {
		t.Fatal("loop need > 0")
	}
}
