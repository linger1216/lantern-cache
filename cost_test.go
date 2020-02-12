package lantern

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCoster(t *testing.T) {
	c := newCoster(10)
	c.add(1, 1)
	require.Equal(t, c.remain(0), int64(9))
	c.add(2, 1)
	require.Equal(t, c.remain(0), int64(8))
	c.add(3, 1)
	require.Equal(t, c.remain(0), int64(7))
	c.updateIfHas(3, 2)
	require.Equal(t, c.remain(0), int64(6))
	c.updateIfHas(4, 1)
	require.Equal(t, c.remain(0), int64(6))
	c.del(2)
	require.Equal(t, c.remain(0), int64(7))
	sample := c.randomPair()
	for i := range sample {
		if i <= 1 {
			require.Greater(t, sample[i].hash, uint64(0))
			require.Greater(t, sample[i].cost, int64(0))
		} else {
			require.Equal(t, sample[i].hash, uint64(0))
			require.Equal(t, sample[i].cost, int64(0))
		}
	}
}