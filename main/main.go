package main

import (
	"bytes"
	lantern_cache "gitee.com/jellyfish/lantern-cache"
)

func blob(char byte, len int) []byte {
	b := make([]byte, len)
	for index := range b {
		b[index] = char
	}
	return b
}

func main() {
	//var message = blob('a', 256)

	//cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
	//	BucketCount: 512,
	//	MaxCapacity: 1024 * 1024,
	//})
	//
	//N := 200000
	//
	//for i := 0; i < N; i++ {
	//	key := []byte(strconv.Itoa(i))
	//	cache.Put(key, blob('a', lantern_cache.RandomNumber(1, 512)))
	//	//cache.Put(key, message)
	//}
	//
	//wg := sync.WaitGroup{}
	//for i := 0; i < 8; i++ {
	//	wg.Add(1)
	//	go func() {
	//		for i := 0; i < N; i++ {
	//			_, err := cache.Get(strconv.Itoa(rand.Intn(N)))
	//			if err != nil && err != lantern_cache.ErrorNotFound {
	//				panic(err)
	//			}
	//		}
	//		wg.Done()
	//	}()
	//}
	//wg.Wait()
	//fmt.Printf("%s\n", cache.Stats().String())
	//fmt.Printf("%s\n", cache.Stats().Raw())

	b := lantern_cache.NewLanternCache(&lantern_cache.Config{
		ChunkAllocatorPolicy: "mmap",
		BucketCount:          32,
		MaxCapacity:          1024 * 1024 * 1024,
	})
	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.Put(key1, val1)
	if err != nil {
		panic(err)
	}
	actual, err := b.Get(key1)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(actual, val1) {
		panic("not equal")
	}
}
