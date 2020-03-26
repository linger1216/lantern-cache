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

	HashFnv = "fnv"
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
	key        string
	value      interface{}
	expiration time.Time
}

type bigEntry struct {
	entry    *entry
	hashed   uint64
	conflict uint64
	cost     int64
}

type OnEvictFunc func(key string)
type CostFunc func(value interface{}) (cost int64)

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
