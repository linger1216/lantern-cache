package lantern

import (
	"testing"
	"time"
)

func TestNewLanternCache(t *testing.T) {

}

func TestCache_SetWithTTL(t *testing.T) {
	c := NewLanternCache(&Config{
		Shards:  16,
		MaxCost: 100,
		OnEvict: func(u uint64, u2 uint64, i interface{}, i2 int64) {

		},
		Hash: "fnvxx",
	})

	for i := 0; i < 1e7; i++ {
		v := randomString(randomNumber(1, 16))
		err := c.SetWithTTL(randomString(randomNumber(1, 16)), v, int64(len(v)), time.Hour)
		if err != nil {
			t.Fatal(err)
		}
	}

}
