package main

import (
	"fmt"
	lantern_cache "github.com/linger1216/lantern-cache"
	"github.com/pkg/profile"
	"math/rand"
	"runtime"
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
	return rand.Intn(max) + min
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// usage()
	runCacheBenchmark(1<<24, 100)
}

func usage() {
	N := 200000
	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
		BucketCount:  512,
		MaxCapacity:  1024 * 1024 * 40,
		InitCapacity: 1024 * 1024 * 5,
	})

	fillBuf := blob('a', RandomNumber(1, 512))

	for i := 0; i < N; i++ {
		err := cache.Put([]byte(strconv.Itoa(i)), fillBuf)
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

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func byteWithCharset(length int, charset string) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return b
}

func randomByte(length int) []byte {
	return byteWithCharset(length, charset)
}

func runCacheBenchmark(N int, pctWrites uint64) {

	defer profile.Start(profile.MemProfile).Stop()

	rc := uint64(0)
	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
		ChunkAllocatorPolicy: "heap",
		BucketCount:          512,
		MaxCapacity:          1024 * 1024 * 500,
		InitCapacity:         1024 * 1024 * 100,
	})

	for i := 0; i < N; i++ {
		kv := []byte(strconv.Itoa(i))
		err := cache.Put(kv, kv)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("init done wait 5s\n")
	fmt.Printf("init %s\n", cache.String())
	time.Sleep(time.Second * 5)

	i := 0
	roundCount := 1 << 8
	for {
		for i := 0; i < N; i++ {
			mc := atomic.AddUint64(&rc, 1)
			key := []byte(strconv.Itoa(rand.Intn(N)))
			if pctWrites*mc/100 != pctWrites*(mc-1)/100 {
				err := cache.Put(key, randomByte(RandomNumber(1, 32)))
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
		i++
		fmt.Printf("round %d %s\n", i, cache.String())
		runtime.GC()
		if i == roundCount {
			break
		}
	}
}

//func printMemStats()  {
//	mem := &runtime.MemStats{}
//	runtime.ReadMemStats(mem)
//	fmt.Printf("已分配的对象的字节数:%d\n", mem.Alloc)
//	fmt.Printf("堆分配的对象的字节数:%d\n", mem.HeapAlloc)
//	fmt.Printf("堆已有的最大尺寸:%d\n", mem.HeapSys)
//	fmt.Printf("堆未被使用的:%d\n", mem.HeapIdle)
//	fmt.Printf("堆使用的:%d\n", mem.HeapInuse)
//	fmt.Printf("堆idle中返还给os:%d\n", mem.HeapReleased)
//	fmt.Printf("栈使用的:%d\n", mem.StackInuse)
//}
