package lantern

import "sync"

type storeShared struct {
	mask   uint64
	shards []*mutexMap
}

func newStoreShared(n uint64) *storeShared {
	ret := &storeShared{}
	if n <= 0 {
		n = 256
	}

	if !isPowerOfTwo(n) {
		panic("shard count need is power of two")
	}

	shards := make([]*mutexMap, n)
	for i := uint64(0); i < n; i++ {
		shards[i] = newMutexMap()
	}
	ret.shards = shards
	ret.mask = n - 1
	return ret
}

func (s *storeShared) Put(entry *entry) error {
	return s.shards[entry.key&s.mask].Put(entry)
}

func (s *storeShared) Get(key, conflict uint64) (interface{}, error) {
	return s.shards[key&s.mask].Get(key, conflict)
}

func (s *storeShared) Del(key, conflict uint64) (uint64, interface{}) {
	return s.shards[key&s.mask].Del(key, conflict)
}

func (s *storeShared) Clean(policy *defaultPolicy, onEvict onEvictFunc) {

}

type mutexMap struct {
	sync.RWMutex
	m map[uint64]*entry
}

func newMutexMap() *mutexMap {
	ret := &mutexMap{}
	ret.m = make(map[uint64]*entry)
	return ret
}

func (s *mutexMap) Put(entry *entry) error {
	s.Lock()
	defer s.Unlock()
	s.m[entry.key] = entry
	return nil
}

func (s *mutexMap) Get(key, conflict uint64) (interface{}, error) {
	s.Lock()
	defer s.Unlock()

	if v, ok := s.m[key]; ok && v.conflict == conflict {
		return v.value, nil
	}
	return nil, ErrorNoEntry
}

func (s *mutexMap) Del(key, conflict uint64) (uint64, interface{}) {
	s.Lock()
	defer s.Unlock()

	entry, ok := s.m[key]
	if !ok {
		return 0, nil
	}

	if conflict != 0 && entry.conflict != conflict {
		return 0, nil
	}

	delete(s.m, key)
	return entry.conflict, entry.value
}
