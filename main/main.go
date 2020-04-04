//package main
//
//import (
//	"fmt"
//	lantern_cache "github.com/linger1216/lantern-cache"
//	"github.com/pkg/profile"
//	runstats "github.com/tevjef/go-runtime-metrics"
//	"math/rand"
//	"runtime"
//	"strconv"
//	"sync"
//	"sync/atomic"
//	"time"
//)
//
//const (
//	KeyCount   = 1 << 24
//	KeySize    = 32
//	ValCount   = 4096
//	ValSize    = 256
//	RoundCount = 1 << 4
//)
//
//var (
//	keys   [][]byte
//	vals   [][]byte
//	cpuNum int
//)
//
//func randomNumber(min, max int) int {
//	return rand.Intn(max) + min
//}
//
//func blob(char byte, len int) []byte {
//	b := make([]byte, len)
//	for index := range b {
//		b[index] = char
//	}
//	return b
//}
//
//func keysList(count, l int) [][]byte {
//	keys := make([][]byte, count)
//	for i := 0; i < count; i++ {
//		b := make([]byte, 0, l)
//		s := l - len(strconv.Itoa(i))
//		b = append(b, []byte(strconv.Itoa(i))...)
//		for i := 0; i < s; i++ {
//			b = append(b, 'a')
//		}
//		keys[i] = b
//	}
//	return keys
//}
//
//func valsList(count, l int) [][]byte {
//	keys := make([][]byte, count)
//	for i := 0; i < count; i++ {
//		keys[i] = blob('a', l)
//	}
//	return keys
//}
//
//func init() {
//	rand.Seed(time.Now().UnixNano())
//	keys = keysList(KeyCount, KeySize)
//	vals = valsList(ValCount, ValSize)
//	cpuNum = runtime.NumCPU()
//}
//
//func main() {
//	usage()
//	// runCacheBenchmark(1<<24, 25)
//}
//
//func runCacheBenchmark(N int, pctWrites uint64) {
//	defer profile.Start(profile.MemProfile).Stop()
//
//	rc := uint64(0)
//	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
//		ChunkAllocatorPolicy: "heap",
//		BucketCount:          512,
//		MaxCapacity:          1024 * 1024 * 1024,
//		InitCapacity:         1024 * 1024 * 100,
//	})
//
//	for i := 0; i < N; i++ {
//		err := cache.Put(keys[randomNumber(0, KeyCount)], vals[randomNumber(0, ValCount)])
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	fmt.Printf("init done.\n")
//
//	runstats.DefaultConfig.CollectionInterval = time.Second
//	if err := runstats.RunCollector(runstats.DefaultConfig); err != nil {
//	}
//
//	i := 0
//	buf := make([]byte, 0, ValSize)
//	for {
//		wg := sync.WaitGroup{}
//		for i := 0; i < cpuNum; i++ {
//			wg.Add(1)
//			go func() {
//				for i := 0; i < N; i++ {
//					mc := atomic.AddUint64(&rc, 1)
//					if pctWrites*mc/100 != pctWrites*(mc-1)/100 {
//						err := cache.Put(keys[randomNumber(0, KeyCount)], vals[randomNumber(0, ValCount)])
//						if err != nil {
//							panic(err)
//						}
//					} else {
//						_, err := cache.GetWithBuffer(buf, keys[randomNumber(0, KeyCount)])
//						if err != nil && err != lantern_cache.ErrorNotFound && err != lantern_cache.ErrorValueExpire {
//							panic(err)
//						}
//					}
//				}
//				wg.Done()
//			}()
//		}
//		wg.Wait()
//		i++
//		fmt.Printf("round %d %s\n", i, cache.String())
//		if i == RoundCount {
//			break
//		}
//	}
//	fmt.Printf("job finished for watch wait 3 mins\n")
//	time.Sleep(time.Second * 180)
//}
//
//func usage() {
//	cache := lantern_cache.NewLanternCache(&lantern_cache.Config{
//		BucketCount:  512,
//		MaxCapacity:  1024 * 1024 * 40,
//		InitCapacity: 1024 * 1024 * 5,
//	})
//
//	err := cache.Put([]byte("hello"), []byte("china"))
//	if err != nil {
//		panic(err)
//	}
//	_, err = cache.Get([]byte("world"))
//	if err != nil && err != lantern_cache.ErrorNotFound && err != lantern_cache.ErrorValueExpire {
//		panic(err)
//	}
//	cache.Reset()
//}

package main

import (
	"github.com/tidwall/redcon"
	"log"
	"strings"
	"sync"
)

var addr = ":6380"

func main() {
	var mu sync.RWMutex
	var items = make(map[string][]byte)
	go log.Printf("started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				mu.Lock()
				items[string(cmd.Args[1])] = cmd.Args[2]
				mu.Unlock()
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				mu.RLock()
				val, ok := items[string(cmd.Args[1])]
				mu.RUnlock()
				if !ok {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
			case "del":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				mu.Lock()
				_, ok := items[string(cmd.Args[1])]
				delete(items, string(cmd.Args[1]))
				mu.Unlock()
				if !ok {
					conn.WriteInt(0)
				} else {
					conn.WriteInt(1)
				}
			}
		},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
