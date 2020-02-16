package lantern

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewNode(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	require.Equal(t, 8192, len(node.data))
}

func TestNewNodeWrite(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)
}

func TestNewNodeWriteRead(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)

	for {
		_, err := node.read()
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 8192, node.r)
}

func TestNewNodeExpire(t *testing.T) {
	alloc := NewChunkAllocator("heap")
	data, err := alloc.getChunk()
	if err != nil {
		panic(err)
	}
	node := newNode(data)
	for i := 0; i < 1e4; i++ {
		err := node.write(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 8192, node.w)
	require.Equal(t, 0, node.r)

	for {
		v, err := node.read()
		if err != nil {
			break
		}
		_ = v
		n := randomNumber(1, 10)
		if n >= 5 {
			node.expire()
		}
	}

	require.Equal(t, node.r, node.w)
	if node.r == len(node.data) {
		return
	}

	for i := node.r; i < len(node.data); i++ {
		v := node.data[i]
		require.Equal(t, uint64(0), v)
	}
}

func TestExpirationPut(t *testing.T) {
	e := newExpiration(AllocPolicy)
	for i := 0; i < 1e4; i++ {
		err := e.put(uint64(i))
		if err != nil {
			break
		}
	}
	require.Equal(t, 1, 1)
}
