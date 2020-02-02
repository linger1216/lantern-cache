package lantern_cache

import (
	"bytes"
	"testing"
	"time"
)

func TestWrap(t *testing.T) {

	ts := time.Now().Unix()
	key1 := []byte("key1")
	value1 := []byte("value1")
	blob := wrapEntry(nil, ts, key1, value1)

	if readTimeStamp(blob) != ts {
		t.Fatalf("except:%d actual:%d", ts, readTimeStamp(blob))
	}

	key := readKey(blob)
	if !bytes.Equal(key, key1) {
		t.Fatalf("except:%s actual:%s", key1, key)
	}

	val := readValue(blob, uint16(len(key)))
	if !bytes.Equal(val, value1) {
		t.Fatalf("except:%s actual:%s", bytes2str(value1), bytes2str(val))
	}
}
