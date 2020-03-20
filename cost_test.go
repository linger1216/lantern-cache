package lantern

//func TestCoster(t *testing.T) {
//	c := newCoster(10)
//	c.put(1, 1)
//	require.Equal(t, c.remain(0), int64(9))
//	c.put(2, 1)
//	require.Equal(t, c.remain(0), int64(8))
//	c.put(3, 1)
//	require.Equal(t, c.remain(0), int64(7))
//	c.updateIfNotExist(3, 2)
//	require.Equal(t, c.remain(0), int64(6))
//	c.updateIfNotExist(4, 1)
//	require.Equal(t, c.remain(0), int64(6))
//	c.del(2)
//	require.Equal(t, c.remain(0), int64(7))
//
//	sample := make([]costerPair, 0, SampleCount)
//	sample = c.fillSample(sample)
//	for i := range sample {
//		if i <= 1 {
//			require.Greater(t, sample[i].hash, uint64(0))
//			require.Greater(t, sample[i].cost, int64(0))
//		} else {
//			require.Equal(t, sample[i].hash, uint64(0))
//			require.Equal(t, sample[i].cost, int64(0))
//		}
//	}
//}
