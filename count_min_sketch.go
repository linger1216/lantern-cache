package lantern

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	HashRound        = 4
	MaxFrequentValue = 0x0f
)

// 这代表一行
type countMinRow []byte

// size要保证是2的幂
func newCountMinRow(size uint64) countMinRow {
	if !isPowerOfTwo(size) {
		panic(ErrorNotPowerOfTwo)
	}
	return make([]byte, size>>1)
}

func (c countMinRow) increment(n uint64) {
	// 1 找到这个值属于哪个byte
	idx := n >> 1
	// 2 确定这个数是奇数还是偶数, 如果是奇数高四位, 偶数低四位
	// 确定奇数只要x & 1 = 1 奇数/ 否则偶数
	// 奇数4, 偶数0
	shift := (n & 1) * 4
	// 奇/偶数都适用
	// 高四位:xxxxyyyy >> 4 => 0000xxxx => &0x0f = xxxx
	// 低四位:xxxxyyyy >> 0 => xxxxyyyy => &0x0f = yyyy
	v := (c[idx] >> shift) & 0x0f
	if v >= MaxFrequentValue {
		return
	}
	// xxxxyyyy + 00010000
	c[idx] += 1 << shift
}

func (c countMinRow) get(index uint64) uint8 {
	// index/2找到具体的byte
	// 根据奇偶找合适的值
	// return byte(r[n/2]>>((n&1)*4)) & 0x0f
	h, l := c.getUnit(c[index>>1])
	if index&1 == 1 {
		return h
	}
	return l
}

// 衰减
// 防止某些短时间过热的词长期在count中, 每隔一段时间将其衰减
func (c countMinRow) decay() {
	for i := range c {
		c[i] = c[i] >> 2
	}
}

func (c countMinRow) reset() {
	for i := range c {
		c[i] = 0
	}
}

// 将每个值都打印出来
func (c countMinRow) string() string {
	ret := " countMinRow dump === "
	size := uint64(len(c) << 1)
	for i := uint64(0); i < size; i++ {
		ret += fmt.Sprintf("[%d]:%d ", i, c.get(i))
	}
	ret += fmt.Sprintf(" ===")
	return ret
}

// 把一个字节的频率分别取出
func (c countMinRow) getUnit(n byte) (uint8, uint8) {
	return n >> 4, n & 0x0f
}

type countMinSketch struct {
	rows  [HashRound]countMinRow
	seeds [HashRound]uint64
	mask  uint64
}

// key最多的条数, 比如你设定100w是你的最大存储数量, 那么就设置1e6
func newCountMinSketch(n uint64) *countMinSketch {
	if n == 0 {
		n = 64
	}
	n = next2Power(n)
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := &countMinSketch{}
	ret.mask = n - 1
	for i := 0; i < HashRound; i++ {
		ret.rows[i] = newCountMinRow(n)
		ret.seeds[i] = source.Uint64()
	}
	return ret
}

// 得出新的hash值
func (c *countMinSketch) toIndex(index int, hashed uint64) uint64 {
	return (hashed ^ c.seeds[index]) & c.mask
}

// 用^操作代表多次hash运算
func (c *countMinSketch) increment(hashed uint64) {
	//fmt.Printf("countMinSketch hashed %d\n", hashed)
	for i := 0; i < HashRound; i++ {
		index := c.toIndex(i, hashed)
		//fmt.Printf("rows[%d]->%d, ", i, index)
		c.rows[i].increment(index)
	}
	//fmt.Printf("\n")
	//for i := 0; i < HashRound; i++ {
	//	fmt.Printf("rows[%d] %s\n", i, c.rows[i].string())
	//}
}

func (c *countMinSketch) estimate(hashed uint64) uint8 {
	frequencies := c.get(hashed)
	var min uint8
	for i := range frequencies {
		if i == 0 {
			min = frequencies[i]
			continue
		}
		if frequencies[i] < min {
			min = frequencies[i]
		}
	}
	return min
}

// 取每个hash算法的频率值, 这时候不进行计算
func (c *countMinSketch) get(hashed uint64) [HashRound]uint8 {
	var ret [HashRound]uint8
	for i := 0; i < HashRound; i++ {
		index := c.toIndex(i, hashed)
		ret[i] = c.rows[i].get(index)
	}
	return ret
}

func (c *countMinSketch) reset() {
	for i := 0; i < HashRound; i++ {
		c.rows[i].reset()
	}
}
