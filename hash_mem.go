package lantern

type memxx struct{}

func newMemXX() hasher {
	return &memxx{}
}

func (f *memxx) hash(key interface{}) (uint64, uint64) {
	if key == nil {
		return 0, 0
	}

	switch k := key.(type) {
	case uint64:
		return k, 0
	case string:
		raw := []byte(k)
		return MemHash(raw), xx(raw)
	case []byte:
		return MemHash(k), xx(k)
	case byte:
		return uint64(k), 0
	case int:
		return uint64(k), 0
	case int32:
		return uint64(k), 0
	case uint32:
		return uint64(k), 0
	case int64:
		return uint64(k), 0
	default:
		panic("Key type not supported")
	}
}
