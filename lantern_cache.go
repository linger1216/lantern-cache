package lantern

type LanternCache struct {
}

type Config struct {
	KeyMaxCount uint64
	MaxCost     uint64
	RingBuffer  uint64
	HashFunc    func(key interface{}) (uint64, uint64)
}
