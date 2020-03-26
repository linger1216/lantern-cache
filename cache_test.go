package lantern

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLanternCacheConfig(t *testing.T) {
	c := &Config{
		MaxCost: 1024 * 1024 * 1024,
		AvgCost: 1024,
	}
	c.defaultValue()

	require.Equal(t, c.Shards, uint64(256))
	require.Equal(t, c.MaxKeyCount, uint64(1024*1024*10))
	require.Equal(t, c.BucketInterval, int64(5))
	require.Equal(t, c.MaxAccessRingBuffer, uint64(64))
	require.Equal(t, c.PutEntryBuffer, uint64(32*1024))
	require.Equal(t, c.Hash, HashFnvXX)
	require.Condition(t, func() (success bool) {
		return c.CostFunc != nil
	})
	require.Condition(t, func() (success bool) {
		return c.OnEvict == nil
	})
	require.Equal(t, c.Log, false)
}

func TestNewLanternCacheMultiClose(t *testing.T) {
	c := NewLanternCache(&Config{
		MaxKeyCount: 1024 * 1024 * 10,
		MaxCost:     1024 * 1024 * 1024,
	})
	c.close()
	c.close()
	c.close()
}

func TestCache_PutGet(t *testing.T) {
	c := NewLanternCache(&Config{
		MaxCost: 1024 * 1024 * 1024,
		AvgCost: 1024,
	})

	tests := []struct {
		name     string
		key      string
		val      string
		duration time.Duration // ms

		expectValue interface{}
		expectRet   bool
		wait        time.Duration
	}{
		{name: "normal", key: "key", val: "value", duration: 10 * time.Second,
			expectValue: "value", expectRet: true, wait: 0 * time.Millisecond},
		{name: "过期key", key: "key", val: "value", duration: 1 * time.Millisecond,
			expectValue: nil, expectRet: false, wait: 1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.PutWithTTL(tt.key, tt.val, 0, tt.duration)
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			actualValue, ret := c.Get(tt.key)
			require.Equal(t, actualValue, tt.expectValue)
			require.Equal(t, ret, tt.expectRet)
		})
	}
	//
	//for i := 0; i < 1e9; i++ {
	//	k := randomString(randomNumber(1, 16))
	//	v := randomString(randomNumber(1, 16))
	//	ts := time.Second * time.Duration(randomNumber(1, 100))
	//	b := c.PutWithTTL(k, v, int64(len(v)), ts)
	//	if !b {
	//		t.Fatal("lid")
	//	}
	//	//time.Sleep(time.Millisecond * 10)
	//}
}
