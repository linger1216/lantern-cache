package lantern

import (
	"fmt"
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

	expire := 10 * time.Millisecond

	tests := []struct {
		name     string
		key      string
		val      string
		duration time.Duration // ms

		expectValue interface{}
		expectRet   bool
		wait        time.Duration
	}{
		{name: "put_get", key: "key1", val: "value", duration: 10 * time.Second,
			expectValue: "value", expectRet: true, wait: expire},
		{name: "过期key", key: "key2", val: "value", duration: 1 * time.Millisecond,
			expectValue: nil, expectRet: false, wait: expire},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.PutWithTTL(tt.key, tt.val, 0, tt.duration)
			time.Sleep(tt.wait)

			hashed, _ := c.hasher.hash(tt.key)
			cost, freq := c.policy.get(hashed)
			require.Equal(t, cost, int64(len(tt.val)))
			require.Equal(t, freq, uint64(0))
			actualValue, ret := c.Get(tt.key)
			require.Equal(t, actualValue, tt.expectValue)
			require.Equal(t, ret, tt.expectRet)
		})
	}
}

func TestCache_PutGetRandom(t *testing.T) {
	c := NewLanternCache(&Config{
		MaxCost: 1024 * 1024 * 1024,
		AvgCost: 1024,
		Stats:   true,
	})

	go func() {
		for {
			fmt.Println(c.Stats())
		}
	}()

	expire := 10 * time.Millisecond

	for i := 0; i < 1000; i++ {
		key, val := randomString(16), randomString(16)
		c.Put(key, val, 0)
		time.Sleep(expire)
		hashed, _ := c.hasher.hash(key)
		cost, freq := c.policy.get(hashed)
		require.Equal(t, cost, int64(len(val)))
		require.Equal(t, freq, uint64(0))
		actualValue, ret := c.Get(key)
		require.Equal(t, actualValue, val)
		require.Equal(t, ret, true)
	}
}
