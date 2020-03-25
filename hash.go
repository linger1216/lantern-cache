package lantern

type hasher interface {
	hash(key []byte) uint64
}
