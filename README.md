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
GOROOT=/usr/local/go #gosetup
GOPATH=/Users/lid.guan/Downloads/go_proc #gosetup
/usr/local/go/bin/go test -c -o /private/var/folders/sw/63_3m7kx17v3z8wh5dncz9w00000gn/T/___gobench_github_com_linger1216_lantern_cache github.com/linger1216/lantern-cache #gosetup
/private/var/folders/sw/63_3m7kx17v3z8wh5dncz9w00000gn/T/___gobench_github_com_linger1216_lantern_cache -test.v -test.bench . -test.run ^$ #gosetup
goos: darwin
goarch: amd64
pkg: github.com/linger1216/lantern-cache
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4         	11336731	       149 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	15416744	       130 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	13527022	        94.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4         	10042275	       148 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	12521772	       160 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	12743173	       160 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4        	 8648589	       180 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	 8673067	       176 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[read]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	 7961178	       171 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 3990651	       272 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4      	 9133687	       145 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4     	 9440418	       180 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 4066081	       276 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4      	 8164665	       194 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4     	 8250060	       177 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	 3080474	       575 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4     	 6063841	       221 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[write]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4    	 6195343	       164 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:256-4          	 3805460	       386 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:256-4        	 5628916	       238 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:256-4       	 6325657	       223 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:512-4          	 3485614	       416 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:512-4        	 5293088	       219 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:512-4       	 6024949	       209 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4         	 2999001	       383 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:512_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4       	 4618534	       279 ns/op	       0 B/op	       0 allocs/op
BenchmarkCaches/[mix]_bucket:1024_capacity:1G_alloc:heap_kenLen:32_valLen:1024-4      	 5348940	       292 ns/op	       0 B/op	       0 allocs/op
PASS
```