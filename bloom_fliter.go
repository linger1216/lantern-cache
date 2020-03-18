package lantern

import (
	"fmt"
	"math"
)

// 返回
// size 最小为512
// size 可能不是2^n的形式, 要找到一个比ui64大的size
// exponent 阶数 通过几次左移
func getSize(ui64 uint64) (size uint64, exponent uint64) {
	if ui64 < uint64(512) {
		ui64 = uint64(512)
	}
	size = uint64(1)
	for size < ui64 {
		size <<= 1
		exponent++
	}
	return size, exponent
}

// 纯理论的东西 不懂
// 解释:https://sagi.io/2017/07/bloom-filters-for-the-perplexed/#appendix
func calcSizeByWrongPositives(numEntries, wrongs float64) (uint64, uint64) {
	size := -1 * numEntries * math.Log(wrongs) / math.Pow(float64(0.69314718056), 2)
	locs := math.Ceil(float64(0.69314718056) * size / numEntries)
	return uint64(size), uint64(locs)
}

type bloomFilter struct {
	mask   uint64
	shift  uint64
	round  uint64
	bitset *bitset
}

/*
	// For example, if you expect your cache to hold 1,000,000 items when full,
	// NumCounters should be 10,000,000 (10x). Each counter takes up 4 bits, so
	// keeping 10,000,000 counters would require 5MB of memory.
*/
// 不能超过2^64次方
func newBloomFilter(numberCounter, para float64) *bloomFilter {
	if numberCounter == 0 || para == 0 {
		panic(ErrorBloomFilterInvalidPara)
	}

	// 期待有多少个数量的key
	// rounds 经过几轮的计算
	// 每轮不是通过独立的hash函数实现的, 而是通过seed的算术方法
	// 这种做法是否优劣, 应该背后有数学方法, 超越了我的知识储备
	// todo
	var entries, rounds uint64
	if para < 1 {
		entries, rounds = calcSizeByWrongPositives(numberCounter, para)
	} else {
		entries, rounds = uint64(numberCounter), uint64(para)
	}

	// 这里没用next2Power, 是因为还要拿到exponent值
	size, exponent := getSize(entries)
	if size <= 1 {
		panic(ErrorBloomFilterInvalidSize)
	}

	bitset := newBitset(size)
	if bitset == nil {
		panic(ErrorBitsetInvalid)
	}

	ret := &bloomFilter{
		mask:   size - 1,
		shift:  64 - exponent,
		round:  rounds,
		bitset: bitset,
	}
	return ret
}

func (bl *bloomFilter) highLow(hash uint64) (uint64, uint64) {
	return hash >> bl.shift, hash << bl.shift >> bl.shift
}

// 通过好几轮来计算不同的值
func (bl *bloomFilter) add(hash uint64) {
	fmt.Printf("want to add %d\n", hash)
	h, l := bl.highLow(hash)
	for i := uint64(0); i < bl.round; i++ {
		index := (h + i*l) & bl.mask
		fmt.Printf("round %d set %d\n", i, index)
		bl.bitset.set(index)
	}
}

func (bl *bloomFilter) exist(hash uint64) bool {
	h, l := bl.highLow(hash)
	for i := uint64(0); i < bl.round; i++ {
		if !bl.bitset.has((h + i*l) & bl.mask) {
			return false
		}
	}
	return true
}

func (bl *bloomFilter) addIfNotExist(hash uint64) bool {
	if !bl.exist(hash) {
		bl.add(hash)
		return true
	}
	return false
}

// 当过滤器增加了越来越多的元素
// bitset中的1也就越来越多, 最终bloom认为任何值在里面都存在
// 所以当插入次数等于size的时候, 可以执行reset操作
// todo
// 不过也不好说, 插入次数等于size, 概率上也不一定覆盖满bloom, 但到底是个什么情况, 又是个数学问题
func (bl *bloomFilter) reset() {
	bl.bitset.reset()
}
