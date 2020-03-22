package lantern

import (
	"fmt"
	"sync"
)

type storeShared struct {
	mask            uint64
	shards          []*mutexMap
	storeExpiration *storeExpiration
	onEvict         onEvictFunc
}

func newStoreShared(n uint64, bucketInterval int64, onEvict onEvictFunc) *storeShared {
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
	ret.storeExpiration = newStoreExpiration(bucketInterval, ret.cleanBucket)
	ret.onEvict = onEvict
	return ret
}

func (s *storeShared) cleanBucket(m bucket) {
	for key, conflict := range m {
		fmt.Printf("del key:%d conflict:%d\n", key, conflict)
		entry := s.Del(key, conflict)
		if s.onEvict != nil {
			s.onEvict(key, conflict, entry.value, entry.cost)
		}
	}
}

func (s *storeShared) Put(entry *entry) error {
	return s.shards[entry.key&s.mask].Put(entry)
}

func (s *storeShared) Get(key, conflict uint64) (interface{}, error) {
	return s.shards[key&s.mask].Get(key, conflict)
}

func (s *storeShared) Del(key, conflict uint64) *entry {
	return s.shards[key&s.mask].Del(key, conflict)
}

func (s *storeShared) Clean() {
	fmt.Printf("store clean\n")
	s.storeExpiration.cleanUp()
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

// conflict 0 代表强制删除
func (s *mutexMap) Del(key, conflict uint64) *entry {
	s.Lock()
	defer s.Unlock()

	entry, ok := s.m[key]
	if !ok {
		return nil
	}

	if conflict != 0 && entry.conflict != conflict {
		return nil
	}

	delete(s.m, key)
	return entry
}
