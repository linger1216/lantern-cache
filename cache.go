package lantern

import (
	"fmt"
	"strings"
	"time"
)

type entry struct {
	key        uint64
	conflict   uint64
	value      interface{}
	cost       int64
	expiration time.Time
}

type OnEvictFunc func(key uint64, conflict uint64, value interface{}, cost int64)
type CostFunc func(value interface{}) (cost int64, err error)

type Cache struct {
	policy           *defaultPolicy
	store            *storeShared
	hasher           hasher
	accessRingBuffer *ringPoll
	cleanupTicker    *time.Ticker
	stopChannel      chan struct{}
	putEntryChannel  chan *entry
	onEvict          OnEvictFunc
	costFunc         CostFunc
}

type Config struct {
	// 分片数
	Shards uint64
	// 最大的key数量, 这个值用于频率统计, 淘汰策略
	MaxKeyCount uint64
	// 这个值影响到ttl策略, 表示相差多少秒内的ttl数据会聚在一起
	BucketInterval int64
	// 最大的成本控制, 可以理解成value的长度和
	MaxCost uint64
	// 访问记录的缓冲区大小, 会影响到总容量
	MaxAccessRingBuffer uint64
	// put数据异步缓冲区大小 默认32K
	PutEntryBuffer uint64
	// 被淘汰的回调函数
	OnEvict OnEvictFunc
	// hash算法的选择, 目前仅仅支持fnv-xx
	Hash string
	// 成本计算的回调函数
	CostFunc CostFunc
}

func (c *Config) defaultValue() {
	if c.Shards == 0 {
		c.Shards = 256
	}

	if c.MaxKeyCount == 0 {
		c.MaxKeyCount = c.MaxCost * 10
	}

	if c.MaxKeyCount == 0 {
		c.MaxKeyCount = 1e8
	}

	if c.BucketInterval == 0 {
		c.BucketInterval = 5
	}

	// 访问记录的缓冲区大小, 会影响到总容量, 默认64
	if c.MaxAccessRingBuffer == 0 {
		c.MaxAccessRingBuffer = 64
	}

	// 发送缓冲区
	if c.PutEntryBuffer == 0 {
		c.PutEntryBuffer = 32 * 1024
	}

	if len(c.Hash) == 0 {
		c.Hash = "fnv-xx"
	}

	// 成本函数
	if c.CostFunc == nil {
		c.CostFunc = DefaultCost
	}
}

func DefaultCost(v interface{}) (int64, error) {
	if v == nil {
		return 1, nil
	}

	switch k := v.(type) {
	case uint64:
		return 8, nil
	case string:
		return int64(len(k)), nil
	case []byte:
		return int64(len(k)), nil
	case byte:
		return 1, nil
	case int:
		return 4, nil
	case int32:
		return 4, nil
	case uint32:
		return 4, nil
	case int64:
		return 8, nil
	default:
		// 这里有可能会有递归的问题, 不知道帮用户设置默认合不合适
		//buf, err := jsoniter.ConfigFastest.Marshal(v)
		//if err != nil {
		//	return 0, nil
		//}
		//return int64(len(buf)), nil
	}
	return 1, nil
}

func NewLanternCache(conf *Config) *Cache {

	switch {
	case conf.MaxCost == 0:
		panic("NumCounters can't be zero")
	}

	conf.defaultValue()

	ret := &Cache{}

	// todo
	// need more test for this value
	switch strings.ToLower(conf.Hash) {
	case "fnv-xx":
		fallthrough
	default:
		ret.hasher = newFnvXX()
	}

	ret.policy = newDefaultPolicy(conf.MaxKeyCount, conf.MaxCost)
	ret.store = newStoreShared(conf.Shards, conf.BucketInterval, conf.OnEvict)
	ret.accessRingBuffer = newRingPool(ret.policy, conf.MaxAccessRingBuffer)
	ret.putEntryChannel = make(chan *entry, conf.PutEntryBuffer)
	ret.onEvict = conf.OnEvict
	ret.cleanupTicker = time.NewTicker(time.Duration(conf.BucketInterval) * time.Second / 2)
	ret.stopChannel = make(chan struct{})

	go ret.process()
	return ret
}

func (c *Cache) close() {
	c.cleanupTicker.Stop()
	c.stopChannel <- struct{}{}
	close(c.stopChannel)
	close(c.putEntryChannel)
	c.policy.close()
}

func (c *Cache) process() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.store.Clean()
		case <-c.stopChannel:
			return
		case entry := <-c.putEntryChannel:
			if entry.cost == 0 && c.costFunc != nil {
				cost, err := c.costFunc(entry.value)
				if err != nil {
					fmt.Printf("err:%s\n", err.Error())
					break
				}
				entry.cost = cost
			}

			evicts, saved, err := c.policy.put(entry.key, entry.cost)
			if err != nil {
				fmt.Printf("err:%s\n", err.Error())
				break
			}

			if saved {
				c.store.Put(entry)
			}

			// 这些都是被淘汰的key, 已经在policy和coster删除掉了
			// 这里要在del中强制删除
			// entry 是有可能为空的, 因为在准备删除的时候被clean up自动清洗掉了
			for i := range evicts {
				entry := c.store.Del(evicts[i], 0)
				if entry != nil && c.onEvict != nil {
					c.onEvict(entry.key, entry.conflict, entry.value, entry.cost)
				}
			}
		}
	}
}

func (c *Cache) Get(key interface{}) (interface{}, error) {
	keyHash, conflict := c.hasher.hash(key)
	c.accessRingBuffer.put(keyHash)
	return c.store.Get(keyHash, conflict)
}

func (c *Cache) Put(key, value interface{}, cost int64) error {
	return c.PutWithTTL(key, value, cost, 0)
}

func (c *Cache) PutWithTTL(key, value interface{}, cost int64, ttl time.Duration) error {
	if c == nil || key == nil {
		return ErrorInvalidPara
	}
	if ttl <= 0 {
		return ErrorNoExpiration
	}
	keyHash, conflict := c.hasher.hash(key)
	entry := &entry{
		key:        keyHash,
		conflict:   conflict,
		value:      value,
		cost:       cost,
		expiration: time.Now().Add(ttl),
	}
	c.putEntryChannel <- entry
	return nil
}
