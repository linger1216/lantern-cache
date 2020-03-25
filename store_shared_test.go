package lantern

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMutexMap_New(t *testing.T) {
	storeExpiration := newStoreExpiration(5, func(b bucket) {})
	m := newMutexMap(storeExpiration)
	_ = m
}

func TestMutexMap_PutGet(t *testing.T) {
	storeExpiration := newStoreExpiration(5, func(b bucket) {})
	m := newMutexMap(storeExpiration)
	tests := []struct {
		name      string
		key       uint64
		conflict1 uint64
		value1    string
		t1        time.Duration

		conflict2   uint64
		expectValue interface{}
		expectErr   error
		wait        time.Duration
	}{
		{name: "putget", key: 1, conflict1: 1, value1: "value", t1: 100,
			conflict2: 1, expectValue: "value", expectErr: nil, wait: 0},
		{name: "过期key", key: 1, conflict1: 1, value1: "value", t1: 1,
			conflict2: 1, expectValue: nil, expectErr: ErrorExpiration, wait: 100},
		{name: "key相同conflict不同", key: 1, conflict1: 1, value1: "value", t1: 100,
			conflict2: 2, expectValue: nil, expectErr: ErrorNoEntry, wait: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Put(&entry{hashed: tt.key, conflict: tt.conflict1, value: tt.value1, cost: int64(len(tt.value1)), expiration: time.Now().Add(tt.t1 * time.Millisecond)})
			time.Sleep(tt.wait * time.Millisecond)
			actualValue, err := m.Get(tt.key, tt.conflict2)
			require.Equal(t, actualValue, tt.expectValue)
			require.Equal(t, err, tt.expectErr)
		})
	}
}

func TestMutexMap_PutPutGet(t *testing.T) {
	storeExpiration := newStoreExpiration(5, func(b bucket) {})
	m := newMutexMap(storeExpiration)
	m.Put(&entry{hashed: 1, conflict: 1, value: "val1", cost: 5, expiration: time.Now().Add(60 * time.Millisecond)})
	m.Put(&entry{hashed: 1, conflict: 2, value: "val2", cost: 5, expiration: time.Now().Add(180 * time.Millisecond)})
	{
		actualValue, err := m.Get(1, 1)
		require.Equal(t, actualValue, "val1")
		require.Equal(t, err, nil)
	}
	{
		actualValue, err := m.Get(1, 2)
		require.Equal(t, actualValue, "val2")
		require.Equal(t, err, nil)
	}
}
