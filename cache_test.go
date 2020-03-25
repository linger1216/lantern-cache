package lantern

import (
	"testing"
)

func TestNewLanternCache(t *testing.T) {
	c := NewLanternCache(&Config{
		MaxKeyCount: 1024 * 1024 * 10,
		MaxCost:     1024 * 1024 * 1024,
	})

	c.close()
}

//func TestCache_SetWithTTL(t *testing.T) {
//	c := NewLanternCache(&Config{
//		Shards:              16,
//		MaxKeyCount:         10000,
//		BucketInterval:      5,
//		MaxCost:             1024,
//		MaxAccessRingBuffer: 64,
//		PutEntryBuffer:      256,
//		OnEvict: func(hashed uint64, conflict uint64, value interface{}, cost int64) {
//			fmt.Printf("[evict] hashed:%d conflict:%d val:%s cost:%d\n", hashed, conflict, value.(string), cost)
//		},
//		Hash:     "fnv-xx",
//		CostFunc: nil,
//	})
//
//	for i := 0; i < 1e7; i++ {
//		k := randomString(randomNumber(1, 16))
//		v := randomString(randomNumber(1, 16))
//		ts := time.Second * time.Duration(randomNumber(1, 100))
//		err := c.PutWithTTL(k, v, int64(len(v)), ts)
//		if err != nil {
//			t.Fatal(err)
//		}
//		time.Sleep(time.Millisecond * 100)
//	}
//}
