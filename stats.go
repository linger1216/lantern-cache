package lantern_cache

import "fmt"

type Stats struct {
	GetCalls   uint64
	PutCalls   uint64
	Errors     uint64
	Hits       uint64
	Misses     uint64
	BytesSize  uint64
	Collisions uint64
}

func (s *Stats) String() string {
	return fmt.Sprintf("hit:%f err:%f collisions:%f cap:%s",
		float32(s.Hits)/float32(s.GetCalls),
		float32(s.Errors)/float32(s.GetCalls+s.PutCalls),
		float32(s.Collisions)/float32(s.GetCalls+s.PutCalls),
		humanSize(int64(s.BytesSize)))
}

func (s *Stats) Raw() string {
	return fmt.Sprintf("get:%d put:%d err:%d hit:%d miss:%d mem:%s collisions:%d",
		s.GetCalls,
		s.PutCalls,
		s.Errors,
		s.Hits,
		s.Misses,
		humanSize(int64(s.BytesSize)),
		s.Collisions)
}
