package lantern_cache

import (
	"bytes"
	"testing"
	"time"
)

func makeByte(size int) []byte {
	return make([]byte, size)
}

func TestBucketPutGet(t *testing.T) {
	b := newBucket(&bucketConfig{
		64 * 1024 * 2,
		0,
		NewChunkAllocator("heap"),
		&Stats{},
	})
	h := newFowlerNollVoHasher()
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.put(h.Hash(key1), key1, val1, 0)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.get(nil, h.Hash(key1), key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}
}

func TestBucketPutGetExpire(t *testing.T) {
	b := newBucket(&bucketConfig{
		64 * 1024 * 2,
		0,
		NewChunkAllocator("heap"),
		&Stats{},
	})
	h := newFowlerNollVoHasher()
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.put(h.Hash(key1), key1, val1, time.Now().Unix()+1)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.get(nil, h.Hash(key1), key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}

	time.Sleep(time.Second * 2)
	_, err = b.get(nil, h.Hash(key1), key1)
	if err != ErrorValueExpire {
		t.Fatal(err)
	}
}

func TestBucketPutGetSmall(t *testing.T) {
	b := newBucket(&bucketConfig{
		1,
		0,
		NewChunkAllocator("heap"),
		&Stats{},
	})
	h := newFowlerNollVoHasher()
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.put(h.Hash(key1), key1, val1, time.Now().Unix()+1)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.get(nil, h.Hash(key1), key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}
}

func TestCacheBigKeyValue(t *testing.T) {
	b := newBucket(&bucketConfig{
		64 * 1024 * 2,
		0,
		NewChunkAllocator("heap"),
		&Stats{},
	})
	h := newFowlerNollVoHasher()
	key1 := []byte("key1")
	val1 := makeByte(64 * 1024)
	err := b.put(h.Hash(key1), key1, val1, time.Now().Unix()+1)
	if err != ErrorInvalidEntry {
		t.Fatal(err)
	}
}

func TestBucketDel(t *testing.T) {
	b := newBucket(&bucketConfig{
		64 * 1024 * 2,
		0,
		NewChunkAllocator("heap"),
		&Stats{},
	})
	h := newFowlerNollVoHasher()
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.put(h.Hash(key1), key1, val1, 0)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := b.get(nil, h.Hash(key1), key1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, val1) {
		t.Fatal("not equal")
	}

	b.del(h.Hash(key1))
	_, err = b.get(nil, h.Hash(key1), key1)
	if err != ErrorNotFound {
		t.Fatal(err)
	}
}
