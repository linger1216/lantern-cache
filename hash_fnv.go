package lantern

import "github.com/cespare/xxhash"

const (
	// offset64 FNVa offset basis. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	offset64 = 14695981039346656037
	// prime64 FNVa prime value. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	prime64 = 1099511628211
)

type hashFnv struct {
}

func (h *hashFnv) hash(k []byte) uint64 {
	return fnv(k)
}

// Sum64 gets the string and returns its uint64 hash value.
func fnv(key []byte) uint64 {
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}
	return hash
}

type hashXX struct {
}

func (h *hashXX) hash(k []byte) uint64 {
	return xx(k)
}

func xx(key []byte) uint64 {
	return xxhash.Sum64(key)
}

//package lantern
//
//
//import "github.com/cespare/xxhash"
//
//type fnvxx struct{}
//
//// newDefaultHasher returns a new 64-bit FNV-1a Hasher which makes no memory allocations.
//// Its Sum64 method will lay the value out in big-endian byte order.
//// See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function
//func newFnvXX() hasher {
//	return &fnvxx{}
//}
//
//const (
//	// offset64 FNVa offset basis. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
//	offset64 = 14695981039346656037
//	// prime64 FNVa prime value. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
//	prime64 = 1099511628211
//)
//
//func (f *fnvxx) hash(key interface{}) (uint64, uint64) {
//	if key == nil {
//		return 0, 0
//	}
//
//	switch k := key.(type) {
//	case uint64:
//		return k, 0
//	case string:
//		raw := []byte(k)
//		return fnv(raw), xx(raw)
//	case []byte:
//		return fnv(k), xx(k)
//	case byte:
//		return uint64(k), 0
//	case int:
//		return uint64(k), 0
//	case int32:
//		return uint64(k), 0
//	case uint32:
//		return uint64(k), 0
//	case int64:
//		return uint64(k), 0
//	default:
//		panic("Key type not supported")
//	}
//}
//
//
//
//// Sum64 gets the string and returns its uint64 hash value.
//func fnv(key []byte) uint64 {
//	var hash uint64 = offset64
//	for i := 0; i < len(key); i++ {
//		hash ^= uint64(key[i])
//		hash *= prime64
//	}
//	return hash
//}
//
//func xx(key []byte) uint64 {
//	return xxhash.Sum64(key)
//}
//
