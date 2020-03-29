package lantern

import (
	"time"
	"unsafe"
)

const (
	CleanCount = 1 << 24

	chunksPerAlloc            = 1024
	MaxKeySize                = 1 << 16
	MaxValueSize              = 1 << 16 // 64k
	EntryTimeStampFieldSizeOf = 8
	EntryKeyFieldSizeOf       = 2
	EntryValueFieldSizeOf     = 2
	EntryHeadFieldSizeOf      = EntryTimeStampFieldSizeOf + EntryKeyFieldSizeOf + EntryValueFieldSizeOf
	OffsetSizeOf              = 40
	LoopSizeOf                = 64 - OffsetSizeOf

	HashFnvXX = "fnv-xx"
	HashMemXX = "mem-xx"
)

func defaultCost(v interface{}) int64 {
	if v == nil {
		return 1
	}
	switch k := v.(type) {
	case string:
		return int64(len(k))
	case []byte:
		return int64(len(k))
	case byte:
		return int64(unsafe.Sizeof(k))
	case int:
		return int64(unsafe.Sizeof(k))
	case uint:
		return int64(unsafe.Sizeof(k))
	case int32:
		return int64(unsafe.Sizeof(k))
	case uint32:
		return int64(unsafe.Sizeof(k))
	case int64:
		return int64(unsafe.Sizeof(k))
	case uint64:
		return int64(unsafe.Sizeof(k))
	default:
		panic(ErrorUnknowSize)
	}
}

type entry struct {
	// hashed 可以考虑是不是删除
	hashed     uint64
	conflict   uint64
	value      interface{}
	expiration time.Time
}

type bigEntry struct {
	entry *entry
	key   interface{}
	cost  int64
}

type OnEvictFunc func(hashed, conflict uint64, val interface{})
type CostFunc func(value interface{}) (cost int64)

type Stats struct {
	hit   uint64
	miss  uint64
	key   uint64
	evict uint64
	drop  uint64
}
