# lantern-cache

idea from fastcache

i'll write more desc, wait wait

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
```
goos: darwin
goarch: amd64
pkg: github.com/linger1216/lantern-cache
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4         	 6862161	       449 ns/op	     256 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	 6353174	       267 ns/op	     256 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	 9194383	       197 ns/op	     256 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4         	 4177326	       375 ns/op	     512 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	 8217338	       267 ns/op	     512 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	 3091779	       355 ns/op	     512 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4        	 1000000	      1825 ns/op	    1024 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	 1000000	      1668 ns/op	    1024 B/op	       1 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	  977929	      1416 ns/op	    1024 B/op	       1 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 4952995	       269 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	 8832696	       168 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4     	10300470	       129 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 3932490	       310 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	 7189101	       187 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4     	 9894856	       181 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	 3073710	       369 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	 5401651	       242 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4    	 5818298	       226 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4          	 2259996	       497 ns/op	     200 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 3831506	       385 ns/op	     194 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	 2767093	       446 ns/op	     199 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4          	 1483357	      1082 ns/op	     390 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 3328225	       676 ns/op	     399 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	 3006183	       653 ns/op	     401 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4         	  995828	      1870 ns/op	     770 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	 1297626	      1817 ns/op	     774 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	  658286	      1741 ns/op	     772 B/op	       0 allocs/op
PASS
```