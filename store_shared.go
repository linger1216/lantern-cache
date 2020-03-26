package lantern

import (
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
		entry := s.Del(key, conflict)
		if entry != nil && s.onEvict != nil {
			s.onEvict(key, conflict, entry.value)
		}
	}
}

func (s *storeShared) Put(entry *entry) {
	s.shards[entry.hashed&s.mask].Put(entry)
}

func (s *storeShared) Get(hashed, conflict uint64) (interface{}, bool) {
	return s.shards[hashed&s.mask].Get(hashed, conflict)
}

func (s *storeShared) Del(hashed, conflict uint64) *entry {
	return s.shards[hashed&s.mask].Del(hashed, conflict)
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

	toInsert := true
	if currentEntry, founded := s.m[entry.hashed]; founded {
		toInsert = false
		// 用之前的hash去寻找记录
		currentConflict := s.expire.get(entry.hashed, currentEntry.expiration)
		if entry.conflict == currentConflict {
			if !entry.expiration.IsZero() {
				// 新的过期时间存在, 意味着过期时间更新
				s.expire.update(entry.hashed, currentEntry.expiration, entry.conflict, entry.expiration)
			} else {
				// 新的过期时间是0, 意味着没有过期时间
				s.expire.del(entry.hashed, currentEntry.expiration)
			}
		}
	}

	if toInsert {
		if !entry.expiration.IsZero() {
			s.expire.put(entry.hashed, entry.conflict, entry.expiration)
		}
		s.m[entry.hashed] = entry
	}
}

func (s *mutexMap) Get(key, conflict uint64) (interface{}, bool) {
	s.Lock()
	entry, ok := s.m[key]
	s.Unlock()

	if !ok {
		return nil, false
	}

	// todo
	// 等于0的场景, 还要思考
	if conflict != entry.conflict {
		return nil, false
	}

	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (s *mutexMap) Del(key, conflict uint64) *entry {
	s.Lock()
	defer s.Unlock()

	entry, ok := s.m[key]
	if !ok {
		return nil
	}

	// conflict 0 代表强制删除, 给policy预备的
	if conflict != 0 && entry.conflict != conflict {
		return nil
	}

	if !entry.expiration.IsZero() {
		s.expire.del(key, entry.expiration)
	}

	delete(s.m, key)
	return entry
}
