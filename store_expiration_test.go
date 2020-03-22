package lantern

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStoreShared_PutGet(t *testing.T) {
	e := newStoreExpiration(5, func(b bucket) {
		for k, v := range b {
			fmt.Printf("del:%d %d\n", k, v)
		}
	})
	for i := uint64(0); i < 1000; i++ {
		ts := time.Now().Add(time.Second * time.Duration(randomNumber(1, 100)))
		key := i
		conflict := uint64(randomNumber(1, 100))
		e.put(key, conflict, ts)
		actualConflict := e.get(i, ts)
		require.Equal(t, conflict, actualConflict)
	}
}

func TestStoreShared_Del(t *testing.T) {
	e := newStoreExpiration(5, func(b bucket) {
		for k, v := range b {
			fmt.Printf("del:%d %d\n", k, v)
		}
	})
	for i := uint64(0); i < 1000; i++ {
		ts := time.Now().Add(time.Second * time.Duration(randomNumber(1, 100)))
		key := i
		conflict := uint64(randomNumber(1, 100))
		e.put(key, conflict, ts)
		e.del(i, ts)
		actualConflict := e.get(i, ts)
		require.Equal(t, actualConflict, uint64(0))
	}
}

func TestStoreShared_Update(t *testing.T) {
	e := newStoreExpiration(5, func(b bucket) {
		for k, v := range b {
			fmt.Printf("del:%d %d\n", k, v)
		}
	})
	for i := uint64(0); i < 1000; i++ {
		key := i
		ts := time.Now().Add(time.Second * time.Duration(randomNumber(1, 100)))
		conflict := uint64(randomNumber(1, 100))
		e.put(key, conflict, ts)

		newts := time.Now().Add(time.Second * time.Duration(randomNumber(1, 100)))
		newConflict := uint64(randomNumber(1, 100))
		e.update(key, ts, newConflict, newts)
		actualConflict := e.get(i, newts)
		require.Equal(t, actualConflict, newConflict)
	}
}

func TestStoreShared_Clean(t *testing.T) {
	e := newStoreExpiration(5, func(b bucket) {
		for k, v := range b {
			fmt.Printf("del:%d %d\n", k, v)
		}
	})

	end := false
	go func() {
		for !end {
			e.cleanUp()
		}
	}()

	go func() {
		for i := uint64(0); i < 1000; i++ {
			e.put(i, i, time.Now().Add(time.Second*time.Duration(randomNumber(1, 100))))
		}
	}()

	time.Sleep(time.Second * 20)
	end = true
}
