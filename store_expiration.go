package lantern

import (
	"fmt"
	"sync"
	"time"
)

type bucket map[uint64]uint64

type storeExpiration struct {
	sync.RWMutex
	buckets        map[int64]bucket
	bucketInterval int64
	callback       func(bucket)
}

// unit. second
func newStoreExpiration(bucketInterval int64, callback func(bucket)) *storeExpiration {
	return &storeExpiration{
		buckets:        make(map[int64]bucket),
		bucketInterval: bucketInterval,
		callback:       callback,
	}
}

func (s *storeExpiration) _storageBucketIndex(t time.Time) int64 {
	return (t.Unix() / s.bucketInterval) + 1
}

func (s *storeExpiration) _cleanBucketIndex(t time.Time) int64 {
	return s._storageBucketIndex(t) - 1
}

func (s *storeExpiration) _put(key, conflict uint64, expiration time.Time) {
	if expiration.IsZero() {
		return
	}
	storageBucketIndex := s._storageBucketIndex(expiration)
	if _, ok := s.buckets[storageBucketIndex]; !ok {
		s.buckets[storageBucketIndex] = make(bucket)
	}
	s.buckets[storageBucketIndex][key] = conflict
	//fmt.Printf("[put] buckets[%d][%d] = %d %f\n", storageBucketIndex, key, conflict, time.Since(expiration).Seconds())
}

func (s *storeExpiration) put(key, conflict uint64, expiration time.Time) {
	if expiration.IsZero() {
		return
	}
	s.Lock()
	defer s.Unlock()
	s._put(key, conflict, expiration)
}

func (s *storeExpiration) get(key uint64, expiration time.Time) uint64 {
	if expiration.IsZero() {
		return 0
	}

	s.RLock()
	defer s.RUnlock()
	storageBucketIndex := s._storageBucketIndex(expiration)
	if v, ok := s.buckets[storageBucketIndex]; ok {
		return v[key]
	}
	return 0
}

func (s *storeExpiration) del(key uint64, expiration time.Time) {
	if expiration.IsZero() {
		return
	}

	s.Lock()
	defer s.Unlock()
	storageBucketIndex := s._storageBucketIndex(expiration)
	if _, ok := s.buckets[storageBucketIndex]; ok {
		delete(s.buckets[storageBucketIndex], key)
	}
}

func (s *storeExpiration) update(key uint64, expiration time.Time, newConflict uint64, newExpiration time.Time) {
	s.Lock()
	defer s.Unlock()
	if expiration.IsZero() {
		s._put(key, newConflict, newExpiration)
		return
	}
	currentStorageBucketIndex := s._storageBucketIndex(expiration)
	if _, ok := s.buckets[currentStorageBucketIndex][key]; !ok {
		s._put(key, newConflict, newExpiration)
		return
	}
	delete(s.buckets[currentStorageBucketIndex], key)
	s._put(key, newConflict, newExpiration)
}

func (s *storeExpiration) cleanUp() int {
	s.Lock()
	defer s.Unlock()
	cleanBucketIndex := s._cleanBucketIndex(time.Now())
	if m, ok := s.buckets[cleanBucketIndex]; ok {
		fmt.Printf("[clean] buckets[%d]\n", cleanBucketIndex)
		delete(s.buckets, cleanBucketIndex)
		if s.callback != nil {
			s.callback(m)
		}
	}
	return len(s.buckets)
}
