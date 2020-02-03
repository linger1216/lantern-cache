package lantern_cache

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"
)

type bucketConfig struct {
	maxCapacity  uint64
	initCapacity uint64
	chunkAlloc   *chunkAllocator
	statistics   *Stats
}

type bucket struct {
	mutex      sync.RWMutex
	m          map[uint64]uint64
	offset     uint64
	loop       uint32
	chunks     [][]byte
	chunkAlloc *chunkAllocator
	statistics *Stats
}

func newBucket(cfg *bucketConfig) *bucket {
	assert(cfg.maxCapacity > 0, "bucket max capacity need > 0")
	if cfg.initCapacity == 0 {
		cfg.initCapacity = cfg.maxCapacity / 4
	}
	ret := &bucket{}
	ret.statistics = cfg.statistics

	needChunkCount := (cfg.maxCapacity + chunkSize - 1) / chunkSize
	assert(needChunkCount > 0, "max bucket chunk count need > 0")

	initChunkCount := (cfg.initCapacity + chunkSize - 1) / chunkSize
	if initChunkCount == 0 {
		initChunkCount = 1
	}

	ret.chunks = make([][]byte, needChunkCount)
	ret.chunkAlloc = cfg.chunkAlloc
	ret.offset = 0
	ret.loop = 0
	ret.m = make(map[uint64]uint64)

	for i := uint64(0); i < initChunkCount; i++ {
		chunk, err := ret.chunkAlloc.getChunk()
		if err != nil {
			panic(err)
		}
		ret.chunks[i] = chunk
	}

	return ret
}

func (b *bucket) put(keyHash uint64, key, val []byte, expire int64) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	atomic.AddUint64(&b.statistics.Puts, 1)
	entrySize := uint64(EntryHeadFieldSizeOf + len(key) + len(val))
	if len(key) == 0 || len(val) == 0 || len(key) > MaxKeySize || len(val) > MaxValueSize || entrySize > chunkSize {
		atomic.AddUint64(&b.statistics.Errors, 1)
		return ErrorInvalidEntry
	}
	offset := b.offset
	nextOffset := offset + entrySize
	chunkIndex := offset / chunkSize
	nextChunkIndex := nextOffset / chunkSize

	if nextChunkIndex > chunkIndex {
		if int(nextChunkIndex) >= len(b.chunks) {
			b.loop++
			//fmt.Printf("chunk(%v) need loop:%d offset:%d nextOffset:%d chunkIndex:%d nextChunkIndex:%d len(b.chunks):%d\n", &b, b.loop, offset, nextOffset, chunkIndex, nextChunkIndex, len(b.chunks))
			chunkIndex = 0
			offset = 0
		} else {
			//b.logger.Printf("bucket chunk[%d] no space to write so jump next chunk[%d] continue loop:%d", chunkIndex, nextChunkIndex, b.loop)
			chunkIndex = nextChunkIndex
			offset = chunkIndex * chunkSize
		}
		nextOffset = offset + entrySize
	}

	if b.chunks[chunkIndex] == nil {
		chunk, err := b.chunkAlloc.getChunk()
		if err != nil {
			atomic.AddUint64(&b.statistics.Errors, 1)
			return ErrorChunkAlloc
		}
		b.chunks[chunkIndex] = chunk
	}

	chunkOffset := offset & (chunkSize - 1) // or offset % chunkSize
	wrapEntry(b.chunks[chunkIndex][chunkOffset:], expire, key, val)
	b.m[keyHash] = (uint64(b.loop) << OffsetSizeOf) | offset
	b.offset = nextOffset
	//fmt.Printf("[%v] key:%s loop:%d offset:%d", &b, key, b.loop, offset)
	return nil
}

func (b *bucket) get(blob []byte, keyHash uint64, key []byte) ([]byte, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	atomic.AddUint64(&b.statistics.Gets, 1)
	v, ok := b.m[keyHash]
	if !ok {
		atomic.AddUint64(&b.statistics.Misses, 1)
		return nil, ErrorNotFound
	}

	loop := uint32(v >> OffsetSizeOf)
	offset := v & 0x000000ffffffffff

	//b.logger.Printf("[%v] get key:%s loop:%d now loop:%d offset:%d now offset:%d", &b, key, loop, b.loop, offset, b.offset)

	// 1. loop == b.loop && offset < b.offset
	// 这种情况发生在写和读没有发生覆盖的情况下, offset记录的是当时写入的offset, b.offset代表已经写入后的offset(可能多次写)
	// 2.loop+1 == b.loop && offset >= b.offset
	// 这种情况说明, 在写入后, 发生了一次覆盖, 但幸运的是, 覆盖后的值, 没有覆盖到这个key这里
	if loop == b.loop && offset < b.offset || (loop+1 == b.loop && offset >= b.offset) {
		chunkIndex := offset / chunkSize
		if int(chunkIndex) >= len(b.chunks) {
			atomic.AddUint64(&b.statistics.Errors, 1)
			return nil, ErrorChunkIndexOutOfRange
		}

		chunkOffset := offset & (chunkSize - 1) // or offset % chunkSize
		timestamp := readTimeStamp(b.chunks[chunkIndex][chunkOffset:])
		if timestamp > 0 && timestamp < time.Now().Unix() {
			return nil, ErrorValueExpire
		}

		readKey := readKey(b.chunks[chunkIndex][chunkOffset:])
		if !bytes.Equal(readKey, key) {
			atomic.AddUint64(&b.statistics.Collisions, 1)
			return nil, ErrorNotFound
		}
		blob = append(blob, readValue(b.chunks[chunkIndex][chunkOffset:], uint16(len(readKey)))...)
		atomic.AddUint64(&b.statistics.Hits, 1)
		return blob, nil
	}

	atomic.AddUint64(&b.statistics.Misses, 1)
	return nil, ErrorNotFound
}

func (b *bucket) del(keyHash uint64) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	delete(b.m, keyHash)
}

func (b *bucket) reset() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	chunks := b.chunks
	for i := range chunks {
		b.chunkAlloc.putChunk(chunks[i])
		chunks[i] = nil
	}

	bm := b.m
	for k := range bm {
		delete(bm, k)
	}
	b.offset = 0
	b.loop = 0
}

// map len
// map cap
// chunk size
// 理论上chunk最多容量
func (b *bucket) stats() (uint64, uint64, uint64, uint64) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	size := uint64(0)
	for i := range b.chunks {
		if b.chunks[i] != nil {
			size += uint64(len(b.chunks[i]))
		}
	}
	return uint64(len(b.m)), uint64(len(b.m) * 16), size, uint64(len(b.chunks)) * chunkSize
}
