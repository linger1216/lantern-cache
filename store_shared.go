package lantern

import (
	"fmt"
	"sync"
	"time"
)

type storeShared struct {
	mask            uint64
	shards          []*mutexMap
	storeExpiration *storeExpiration
	onEvict         OnEvictFunc
}

func newStoreShared(n uint64, bucketInterval int64, onEvict OnEvictFunc) *storeShared {
	ret := &storeShared{}
	if n <= 0 {
		n = 256
	}

	if !isPowerOfTwo(n) {
		panic("shard count need is power of two")
	}

	shards := make([]*mutexMap, n)
	ret.storeExpiration = newStoreExpiration(bucketInterval, ret.cleanBucket)

	for i := uint64(0); i < n; i++ {
		shards[i] = newMutexMap(ret.storeExpiration)
	}
	ret.shards = shards
	ret.mask = n - 1
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

func (s *storeShared) Put(entry *entry) {
	s.shards[entry.key&s.mask].Put(entry)
}

func (s *storeShared) Get(key, conflict uint64) (interface{}, error) {
	return s.shards[key&s.mask].Get(key, conflict)
}

func (s *storeShared) Del(key, conflict uint64) *entry {
	return s.shards[key&s.mask].Del(key, conflict)
}

func (s *storeShared) Clean() {
	s.storeExpiration.cleanUp()
}

type mutexMap struct {
	sync.RWMutex
	m      map[uint64]*entry
	expire *storeExpiration
}

func newMutexMap(expire *storeExpiration) *mutexMap {
	assert(expire != nil, "expire is null")
	ret := &mutexMap{}
	ret.m = make(map[uint64]*entry)
	ret.expire = expire
	return ret
}

func (s *mutexMap) Put(entry *entry) {
	s.Lock()
	defer s.Unlock()

	currentEntry, ok := s.m[entry.key]
	if !ok {
		if !entry.expiration.IsZero() {
			s.expire.put(entry.key, entry.conflict, entry.expiration)
		}
	} else {
		if entry.conflict == currentEntry.conflict {
			s.expire.update(currentEntry.key, currentEntry.expiration, entry.conflict, entry.expiration)
		} else {
			s.expire.put(entry.key, entry.conflict, entry.expiration)
		}
	}
	s.m[entry.key] = entry
}

func (s *mutexMap) Get(key, conflict uint64) (interface{}, error) {
	s.Lock()
	entry, ok := s.m[key]
	s.Unlock()

	if !ok {
		return nil, ErrorNoEntry
	}

	if conflict != 0 && (conflict != entry.conflict) {
		return nil, ErrorNoEntry
	}

	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		return nil, ErrorExpiration
	}

	return entry.value, nil
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

	if !entry.expiration.IsZero() {
		s.expire.del(key, entry.expiration)
	}
	delete(s.m, key)
	return entry
}
