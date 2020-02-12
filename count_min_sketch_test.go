package lantern

import (
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCountMinRowIncrement(t *testing.T) {
	N := uint64(16)
	cmr := newCountMinRow(N)
	cmr.increment(0)
	cmr.increment(1)
	cmr.increment(0)
	cmr.increment(1)

	require.Equal(t, cmr.get(0), uint8(2))
	require.Equal(t, cmr.get(1), uint8(2))
}

func TestCountMinRowIncrementGet(t *testing.T) {
	N := uint64(16)
	Mask := N - 1
	cmr := newCountMinRow(N)
	R := N * MaxFrequentValue
	m := make(map[uint64]uint64)
	for i := uint64(0); i < R; i++ {
		n := uint64(randomNumber(0, 1000000)) & Mask
		if _, ok := m[n]; ok {
			m[n] += 1
		} else {
			m[n] = 1
		}
		cmr.increment(n)
	}

	for i := uint64(0); i < N; i++ {
		cmrv := cmr.get(i)
		_, ok := m[i]
		require.Equal(t, ok, true)
		mapv := uint8(m[i])
		if mapv >= MaxFrequentValue {
			mapv = MaxFrequentValue
		}
		require.Equal(t, cmrv, mapv)
	}
}

// 每轮都有个单独的数组来存放的, 每次都要用4中算法去拿对应数据
//countMinSketch hashed 446976
//rows[0]->7, rows[1]->4, rows[2]->5, rows[3]->1,
//rows[0]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:0 [5]:0 [6]:0 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:1 [5]:0 [6]:0 [7]:0  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:0 [5]:1 [6]:0 [7]:0  ===
//rows[3]  countMinRow dump === [0]:0 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:0 [7]:0  ===
//countMinSketch hashed 563561
//rows[0]->6, rows[1]->5, rows[2]->4, rows[3]->0,
//rows[0]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:0 [5]:0 [6]:1 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:0 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[3]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:0 [7]:0  ===
//countMinSketch hashed 897822
//rows[0]->1, rows[1]->2, rows[2]->3, rows[3]->7,
//rows[0]  countMinRow dump === [0]:0 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:1 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:0 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:0 [3]:1 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[3]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:0 [7]:1  ===
//countMinSketch hashed 451159
//rows[0]->0, rows[1]->3, rows[2]->2, rows[3]->6,
//rows[0]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:1 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:0 [7]:0  ===
//rows[3]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:0 [5]:0 [6]:1 [7]:1  ===
//countMinSketch hashed 35530
//rows[0]->5, rows[1]->6, rows[2]->7, rows[3]->3,
//rows[0]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:0 [5]:1 [6]:1 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:1 [7]:0  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:0 [7]:1  ===
//rows[3]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:1 [4]:0 [5]:0 [6]:1 [7]:1  ===
//countMinSketch hashed 996747
//rows[0]->4, rows[1]->7, rows[2]->6, rows[3]->2,
//rows[0]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:1 [5]:1 [6]:1 [7]:1  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:1 [7]:1  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:1 [6]:1 [7]:1  ===
//rows[3]  countMinRow dump === [0]:1 [1]:1 [2]:1 [3]:1 [4]:0 [5]:0 [6]:1 [7]:1  ===
//countMinSketch hashed 304216
//rows[0]->7, rows[1]->4, rows[2]->5, rows[3]->1,
//rows[0]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:1 [5]:1 [6]:1 [7]:2  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:2 [5]:1 [6]:1 [7]:1  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:1 [5]:2 [6]:1 [7]:1  ===
//rows[3]  countMinRow dump === [0]:1 [1]:2 [2]:1 [3]:1 [4]:0 [5]:0 [6]:1 [7]:1  ===
//countMinSketch hashed 886585
//rows[0]->6, rows[1]->5, rows[2]->4, rows[3]->0,
//rows[0]  countMinRow dump === [0]:1 [1]:1 [2]:0 [3]:0 [4]:1 [5]:1 [6]:2 [7]:2  ===
//rows[1]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:2 [5]:2 [6]:1 [7]:1  ===
//rows[2]  countMinRow dump === [0]:0 [1]:0 [2]:1 [3]:1 [4]:2 [5]:2 [6]:1 [7]:1  ===
//rows[3]  countMinRow dump === [0]:2 [1]:2 [2]:1 [3]:1 [4]:0 [5]:0 [6]:1 [7]:1  ===
type debugNewHashRound struct {
	newHash uint64
	round   int
}

func TestCountMinSketchIncrementGet(t *testing.T) {
	N := uint64(1 << 7)
	cms := newCountMinSketch(N)
	m := make(map[uint64][4]debugNewHashRound)
	counter := [4]map[uint64]uint64{}
	for i := 0; i < HashRound; i++ {
		counter[i] = make(map[uint64]uint64)
	}
	for i := uint64(0); i < N; i++ {
		hashed := uint64(randomNumber(0, 1000000))
		var newHashArr [4]debugNewHashRound
		for j := 0; j < HashRound; j++ {
			newHashed := cms.toIndex(j, hashed)
			newHashArr[j] = debugNewHashRound{newHashed, j}
			if _, ok := counter[j][newHashed]; ok {
				if counter[j][newHashed] < 15 {
					counter[j][newHashed] += 1
				}
			} else {
				counter[j][newHashed] = 1
			}
		}
		m[hashed] = newHashArr
		cms.increment(hashed)
	}

	for k, v := range m {
		min1 := cms.estimate(k)
		min2 := uint8(math.MaxUint8)
		for i := range v {
			newHashed, round := v[i].newHash, v[i].round
			v, ok := counter[round][newHashed]
			if !ok {
				require.Fail(t, "not ok")
			}
			if uint8(v) < min2 {
				min2 = uint8(v)
			}
		}
		require.Equal(t, min1, min2)
	}
}
