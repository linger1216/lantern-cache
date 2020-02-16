package lantern

import "testing"

func TestGetChunk(t *testing.T) {
	ca := NewChunkAllocator("heap")
	chunk, err := ca.getChunk()
	if err != nil || len(chunk) != chunkSize {
		t.Fatal(err)
	}
}

func TestPutChunk(t *testing.T) {
	ca := NewChunkAllocator("heap")
	chunk, err := ca.getChunk()
	if err != nil || len(chunk) != chunkSize {
		t.Fatal(err)
	}
	ca.putChunk(chunk)
}
