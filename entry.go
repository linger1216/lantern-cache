package lantern_cache

import (
	"encoding/binary"
)

/*
┌───────────────────┐
│   entry marshal   │
├─────┬─────┬─────┬─┴───┬─────┐
│  8  │  2  │  2  │  n  │  m  │
│     │     │     │     │     │
├─────┼─────┼─────┼─────┼─────┤
│ ts  │ key │ val │ key │ val │
│     │size │size │     │     │
└─────┴─────┴─────┴─────┴─────┘
*/
func wrapEntry(blob []byte, timestamp int64, key, val []byte) []byte {
	size := EntryHeadFieldSizeOf + len(key) + len(val)
	if blob == nil {
		blob = make([]byte, size)
	}
	assert(cap(blob) >= size, "wrapEntry blob size need bigger than entry marshal")
	pos := 0

	binary.LittleEndian.PutUint64(blob[pos:pos+EntryTimeStampFieldSizeOf], uint64(timestamp))
	pos += EntryTimeStampFieldSizeOf

	binary.LittleEndian.PutUint16(blob[pos:pos+EntryKeyFieldSizeOf], uint16(len(key)))
	pos += EntryKeyFieldSizeOf

	binary.LittleEndian.PutUint16(blob[pos:pos+EntryValueFieldSizeOf], uint16(len(val)))
	pos += EntryValueFieldSizeOf

	copy(blob[pos:], key)
	pos += len(key)

	copy(blob[pos:], val)
	pos += len(val)
	return blob
}

// 返回位置正好是val部分的起始位置
func readKey(blob []byte) []byte {
	pos := EntryTimeStampFieldSizeOf
	keySize := binary.LittleEndian.Uint16(blob[pos : pos+EntryKeyFieldSizeOf])
	pos += EntryKeyFieldSizeOf + EntryValueFieldSizeOf
	return blob[pos : pos+int(keySize)]
}

func readValue(blob []byte, keySize uint16) []byte {
	pos := EntryTimeStampFieldSizeOf + EntryKeyFieldSizeOf
	valueSize := binary.LittleEndian.Uint16(blob[pos : pos+EntryValueFieldSizeOf])
	pos += EntryValueFieldSizeOf + int(keySize)
	return blob[pos : pos+int(valueSize)]
}

func readTimeStamp(blob []byte) int64 {
	pos := 0
	timestamp := binary.LittleEndian.Uint64(blob[pos : pos+EntryTimeStampFieldSizeOf])
	return int64(timestamp)
}
