package lantern_cache

import "fmt"

type Stats struct {
	Gets       uint64
	Puts       uint64
	Errors     uint64
	Hits       uint64
	Misses     uint64
	BytesSize  int64
	Collisions uint64
}

func (s *Stats) String() string {
	return fmt.Sprintf("hit:%f err:%f collisions:%f bytes:%s",
		float32(s.Hits)/float32(s.Gets),
		float32(s.Errors)/float32(s.Gets+s.Puts),
		float32(s.Collisions)/float32(s.Gets+s.Puts),
		humanSize(int64(s.BytesSize)))
}

func (s *Stats) Raw() string {
	return fmt.Sprintf("get:%d put:%d err:%d hit:%d miss:%d bytes:%s collisions:%d",
		s.Gets,
		s.Puts,
		s.Errors,
		s.Hits,
		s.Misses,
		humanSize(int64(s.BytesSize)),
		s.Collisions)
}
