package lantern

import "testing"

func TestNewTinyLFU(t *testing.T) {
	for i := 0; i <= 10; i++ {
		n := uint64(randomNumber(0, 1<<6))
		newTinyLFU(n)
	}
}
