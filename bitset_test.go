package lantern

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SetHas(t *testing.T) {
	N := uint64(1000000)
	b := newBitset(N)
	for i := uint64(0); i < N; i++ {
		b.set(i)
		existed := b.has(i)
		require.Equal(t, existed, true)
	}
}
