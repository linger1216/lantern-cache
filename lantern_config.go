package lantern_cache

type Config struct {
	BucketCount          uint32
	MaxCapacity          uint64
	InitCapacity         uint64
	ChunkAllocatorPolicy string
	HashPolicy           string
}

func DefaultConfig() *Config {
	return &Config{
		BucketCount:          1024,
		MaxCapacity:          1024 * 1024 * 1024,
		InitCapacity:         1024 * 1024 * 100,
		ChunkAllocatorPolicy: "heap",
		HashPolicy:           "fnv",
	}
}
