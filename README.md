# lantern-cache

- max v len 64KB
- thread safe

### usage
```
func main() {
	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
		BucketCount:  512,
		MaxCapacity:  1024 * 1024 * 40,
		InitCapacity: 1024 * 1024 * 5,
	})

	err := cache.Put([]byte("hello"), []byte("china"))
	if err != nil {
		panic(err)
	}
	_, err = cache.Get([]byte("world"))
	if err != nil && err != lantern_cache.ErrorNotFound && err != lantern_cache.ErrorValueExpire {
		panic(err)
	}
	cache.Reset()
}
```


### benchmark
- alloc from heap
```
goos: darwin
goarch: amd64
pkg: github.com/linger1216/lantern-cache
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4         	 7151116	       245 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	 7108485	       238 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	 6221467	       233 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4         	 5946895	       259 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	 6670867	       368 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	 5518003	       332 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4        	 8055612	       189 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	 9266700	       157 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	 5758790	       286 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 2746644	       425 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	 7245910	       211 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4     	 8116939	       237 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 2675181	       576 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	 7471868	       274 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4     	 8697382	       259 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	 2580746	       863 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	 6427698	       314 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4    	 5146905	       351 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4          	 2390143	       580 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 10110430	       176 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	 8525252	       185 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4          	 2541436	       550 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 9533241	       246 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	 9117368	       275 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4         	 2843509	       657 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	10732830	       212 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	 8898570	       249 ns/op	       0 B/op	       0 allocs/op

```


- alloc from mmap
```
pkg: github.com/linger1216/lantern-cache
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           10713336               191 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4          7468760               203 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4         9556656               215 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4           10451970               218 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4          7877008               291 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4         6027376               312 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4           9160233               179 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4         6528870               432 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4        6574810               330 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           2767180               400 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4         8716328               153 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4        9171146               136 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4           2682656               484 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4         7170303               182 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4        7321443               170 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4          2015647               980 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4        6533864               362 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4       5094814               353 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4             2611654               485 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           6365592               220 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4          8879248               199 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4             3433490               631 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4           4773789               215 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4          5198325               231 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4            2575622               660 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4          6118138               206 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4         4726443               269 ns/op               0 B/op          0 allocs/op
PASS
```


### environment
- MacBook Pro (13-inch, 2017, Two Thunderbolt 3 ports)
- 2.3 GHz Intel Core i5
- 8 GB 2133 MHz LPDDR
