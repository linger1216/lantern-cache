package lantern

type hasher interface {
	hash(k interface{}) (uint64, uint64)
}
