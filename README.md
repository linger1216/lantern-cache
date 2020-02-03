# lantern-cache

- max v len 64KB
- thread safe

### usage
```

func blob(char byte, len int) []byte {
	b := make([]byte, len)
	for index := range b {
		b[index] = char
	}
	return b
}

func RandomNumber(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max) + min
}

func main() {
	N := 200000
	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
		BucketCount:  512,
		MaxCapacity:  1024 * 1024 * 40,
		InitCapacity: 1024 * 1024 * 5,
	})

	for i := 0; i < N; i++ {
		err := cache.Put([]byte(strconv.Itoa(i)), blob('a', RandomNumber(1, 512)))
		if err != nil {
			panic(err)
		}
	}

	core := 8
	wg := sync.WaitGroup{}
	for i := 0; i < core; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < N; i++ {
				_, err := cache.Get([]byte(strconv.Itoa(rand.Intn(N))))
				if err != nil && err != lantern_cache.ErrorNotFound && err != lantern_cache.ErrorValueExpire {
					panic(err)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("%s\n", cache.Stats().String())
	fmt.Printf("%s\n", cache.Stats().Raw())
}
```


### benchmark
- alloc from heap
```
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4           12393765               108 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4         16558442               128 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        11881024               117 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4           10948254               125 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4         12483867               162 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        13507720               158 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4           8720259               212 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4         9942254               206 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4        9516438               128 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4           4641710               243 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4         6730789               168 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       11353910               116 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4           4146306               284 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4         9277275               180 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       10435339               162 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4          3529626               399 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4        6396568               217 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       8316360               178 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4             3633373               386 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4           5452452               250 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4          5973034               200 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4             2751213               434 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4           5763490               250 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4          6115072               258 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4            2593638               395 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4          4023894               292 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4         4678814               246 ns/op               0 B/op          0 allocs/op
PASS

```


- alloc from mmap
```
goos: darwin
goarch: amd64
pkg: github.com/linger1216/lantern-cache
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           10848022               144 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4         15294328                82.1 ns/op             0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4        12562880               127 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4            8294994               163 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4         11843563               135 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4        12868899               110 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4           5119826               219 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4         8657952               197 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4        9322738               155 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           4063599               252 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4         8473459               152 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4       10571989               119 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4           3926442               336 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4         7367847               165 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4       10426177               132 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4          3467011               396 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4        7176171               196 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4       7783929               223 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4             3357818               358 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4           6996694               230 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:256-4          5949139               193 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4             3018706               357 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4           5408210               244 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:512-4          5737129               249 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4            2831487               389 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4          4077646               258 ns/op               0 B/op          0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:mmap_kenLen:32_valLen:1024-4         5917294               253 ns/op               0 B/op          0 allocs/op
PASS

```


### memory


Using Heap strategy memory will double, but using the mmap strategy's memory will only grow by a fifth.