package lantern

type store interface {
	Put(entry *entry) error
	Get(key, conflict uint64) (interface{}, error)
	Del(key, conflict uint64) (uint64, interface{})
	Clean(policy *defaultPolicy, onEvict onEvictFunc)
}
