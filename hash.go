package lantern

type hasher interface {
	hash(key interface{}) (uint64, uint64)
}
