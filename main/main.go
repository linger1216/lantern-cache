package main

import (
	"fmt"
	lantern_cache "github.com/linger1216/lantern-cache"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

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
	// usage()
	runCacheBenchmark(1<<20, 25)
}

func usage() {
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

func runCacheBenchmark(N int, pctWrites uint64) {
	rc := uint64(0)
	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
		ChunkAllocatorPolicy: "mmap",
		BucketCount:          1024,
		MaxCapacity:          1024 * 1024 * 500,
		InitCapacity:         1024 * 1024 * 100,
	})

	for i := 0; i < N; i++ {
		err := cache.Put([]byte(strconv.Itoa(i)), blob('a', RandomNumber(1, 256)))
		if err != nil {
			panic(err)
		}
	}

	count := 0
	breakCount := 1 << 4
	for {
		for i := 0; i < N; i++ {
			mc := atomic.AddUint64(&rc, 1)
			key := []byte(strconv.Itoa(rand.Intn(N)))
			if pctWrites*mc/100 != pctWrites*(mc-1)/100 {
				err := cache.Put(key, blob('a', RandomNumber(1, 256)))
				if err != nil {
					panic(err)
				}
			} else {
				_, err := cache.Get(key)
				if err != nil && err != lantern_cache.ErrorNotFound && err != lantern_cache.ErrorValueExpire {
					panic(err)
				}
			}
		}
		count++
		fmt.Printf("round %d %s\n", count, cache.String())
		if count == breakCount {
			break
		}
	}
}
