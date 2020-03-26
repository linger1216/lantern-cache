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
		if s.onEvict != nil {
			s.onEvict(key, conflict, entry.value, entry.cost)
		}
	}
}

func (s *storeShared) Put(hashed, conflict uint64, entry *entry) {
	s.shards[hashed&s.mask].Put(hashed, conflict, entry)
}

func (s *storeShared) Get(hashed, conflict uint64, key string) (interface{}, bool) {
	return s.shards[hashed&s.mask].Get(key)
}

func (s *storeShared) Del(hashed, conflict uint64, key string) {
	s.shards[hashed&s.mask].Del(hashed, conflict, key)
}

func (s *storeShared) Clean() {
	s.storeExpiration.cleanUp()
}

type mutexMap struct {
	sync.RWMutex
	m      map[string]*entry
	expire *storeExpiration
}

func newMutexMap(expire *storeExpiration) *mutexMap {
	assert(expire != nil, "expire is null")
	ret := &mutexMap{}
	ret.m = make(map[string]*entry)
	ret.expire = expire
	return ret
}

func (s *mutexMap) Put(hashed, conflict uint64, entry *entry) {
	s.Lock()
	defer s.Unlock()

	currentEntry, founded := s.m[entry.key]
	if !founded {
		if !entry.expiration.IsZero() {
			// 将key抽象化了hash, conflict
			s.expire.put(hashed, conflict, entry.expiration)
		}
	} else {
		// 用之前的hash去寻找记录
		currentConflict := s.expire.get(hashed, currentEntry.expiration)
		if conflict == currentConflict {
			if !entry.expiration.IsZero() {
				// 新的过期时间是0, 意味着过期时间更新
				s.expire.update(hashed, currentEntry.expiration, conflict, entry.expiration)
			} else {
				// 新的过期时间是0, 意味着没有过期时间
				s.expire.del(hashed, currentEntry.expiration)
			}
		} else {
			// 这里应该是碰到了hash冲突
			// 同样的hashed, 但取出来的entry, conflict不一样.
			// 这里还是采取插入处理, 插入也会引入2种情况:
			// 1: 完美插入, 因为过期时间会决定bucket, 虽然相同的key, 但不会冲突
			// 2: 覆盖, 此时bucket也相同, 这就没办法了 (也就是说此时只能保留一条expiration过期时间)
			if !entry.expiration.IsZero() {
				s.expire.put(hashed, conflict, entry.expiration)
			}
		}
	}
	s.m[entry.key] = entry
}

func (s *mutexMap) Get(key string) (interface{}, bool) {
	s.Lock()
	entry, ok := s.m[key]
	s.Unlock()

	if !ok {
		return nil, false
	}

	// 这里没有检测过期expire存放的conflict是不是和key完全的一致
	// 因为在put时候已经做了处理
	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (s *mutexMap) Del(hashed, conflict uint64, key string) {
	s.Lock()
	defer s.Unlock()

	if currentEntry, founded := s.m[key]; founded {
		if currentConflict := s.expire.get(hashed, currentEntry.expiration); conflict == currentConflict {
			s.expire.del(hashed, currentEntry.expiration)
		}
		// else 如下:
		// 这里应该是碰到了hash冲突
		// 同样的hashed, 但取出来的entry, conflict不一样.
		// 现在是这样的情况, key相同, 但entry, conflict不一样, 应该是不太可能
	}
	delete(s.m, key)
}
