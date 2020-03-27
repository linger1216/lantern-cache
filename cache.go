package lantern

import (
	"strings"
	"time"
)

type Config struct {
	// 分片数
	Shards uint64
	// 最大的key数量, 这个值用于频率统计, 淘汰策略
	MaxKeyCount uint64

	// 平均成本, 用于辅助计算key数量
	AvgCost uint64

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
	// debug所用
	Log bool
}

func (c *Config) defaultValue() {
	if c.Shards == 0 {
		c.Shards = 256
	}

	if c.MaxKeyCount == 0 && c.AvgCost != 0 {
		c.MaxKeyCount = (c.MaxCost / c.AvgCost) * 10
	}

	if c.BucketInterval == 0 {
		c.BucketInterval = 5
	}

	// 访问记录的缓冲区大小, 会影响到总容量, 默认64
	if c.MaxAccessRingBuffer == 0 {
		c.MaxAccessRingBuffer = 64
	}

	// 发送缓冲区 32k
	if c.PutEntryBuffer == 0 {
		c.PutEntryBuffer = 32 * 1024
	}

	if len(c.Hash) == 0 {
		c.Hash = HashFnvXX
	}

	// 成本函数
	if c.CostFunc == nil {
		c.CostFunc = defaultCost
	}
}

type Cache struct {
	policy           *defaultPolicy
	store            *storeShared
	hasher           hasher
	accessRingBuffer *ringPoll
	cleanupTicker    *time.Ticker
	stopChannel      chan struct{}
	putEntryChannel  chan *bigEntry
	onEvict          OnEvictFunc
	costFunc         CostFunc
	logger           Logger
}

func NewLanternCache(conf *Config) *Cache {

	switch {
	case conf.MaxCost == 0:
		panic("MaxCost can't be zero")
	case conf.AvgCost == 0:
		panic("AvgCost can't be zero")
	}
	conf.defaultValue()

	c := &Cache{}
	switch strings.ToLower(conf.Hash) {
	case HashFnvXX:
		c.hasher = newFnvXX()
	case HashMemXX:
		c.hasher = newMemXX()
	default:
		c.hasher = newFnvXX()
	}

	c.policy = newDefaultPolicy(conf.MaxKeyCount, conf.MaxCost)
	c.store = newStoreShared(conf.Shards, conf.BucketInterval, conf.OnEvict)
	c.accessRingBuffer = newRingPool(c.policy, conf.MaxAccessRingBuffer)
	c.putEntryChannel = make(chan *bigEntry, conf.PutEntryBuffer)
	c.cleanupTicker = time.NewTicker(time.Duration(conf.BucketInterval) * time.Second / 2)
	c.stopChannel = make(chan struct{})

	c.onEvict = conf.OnEvict
	c.costFunc = conf.CostFunc

	if conf.Log {
		c.logger = DefaultLogger()
	} else {
		c.logger = NoneLogger()
	}

	ensure(c.policy != nil)
	ensure(c.store != nil)
	ensure(c.hasher != nil)
	ensure(c.accessRingBuffer != nil)
	ensure(c.cleanupTicker != nil)
	ensure(c.putEntryChannel != nil)
	ensure(c.costFunc != nil)
	ensure(c.logger != nil)

	go c.process()
	return c
}

func (c *Cache) close() {
	if c == nil || c.stopChannel == nil {
		return
	}
	c.cleanupTicker.Stop()
	c.stopChannel <- struct{}{}
	close(c.stopChannel)
	close(c.putEntryChannel)
	c.policy.close()
	c.stopChannel = nil
}

func (c *Cache) process() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.store.Clean()
		case <-c.stopChannel:
			return
		case bigEntry := <-c.putEntryChannel:
			c._save(bigEntry)
		}
	}
}

func (c *Cache) _save(bigEntry *bigEntry) {
	bigEntry.entry.hashed, bigEntry.entry.conflict = c.hasher.hash(bigEntry.key)
	if bigEntry.cost <= 0 {
		bigEntry.cost = c.costFunc(bigEntry.entry.value)
	}
	evicts, saved, err := c.policy.put(bigEntry.entry.hashed, bigEntry.cost)
	if err != nil {
		c.logger.Printf("policy put err:%s\n", err.Error())
		return
	}

	if saved {
		c.store.Put(bigEntry.entry)
	}

	// 这些都是被淘汰的key, 已经在policy和coster删除掉了
	// 这里要在del中强制删除
	// bigEntry 是有可能为空的, 因为在准备删除的时候被clean up自动清洗掉了
	for i := range evicts {
		entry := c.store.Del(evicts[i], 0)
		if entry != nil && c.onEvict != nil {
			c.onEvict(entry.hashed, entry.conflict, entry.value)
		}
	}
}

func (c *Cache) Get(key interface{}) (interface{}, bool) {
	keyHash, conflict := c.hasher.hash(key)
	c.accessRingBuffer.put(keyHash)
	return c.store.Get(keyHash, conflict)
}

func (c *Cache) Put(key interface{}, value interface{}, cost int64) bool {
	return c.PutWithTTL(key, value, cost, 0)
}

func (c *Cache) PutWithTTL(key interface{}, value interface{}, cost int64, ttl time.Duration) bool {
	if c == nil || key == nil || ttl < 0 || cost < 0 {
		return false
	}

	var expiration time.Time
	switch {
	case ttl == 0:
		break
	case ttl < 0:
		return false
	default:
		expiration = time.Now().Add(ttl)
	}

	entry := &bigEntry{
		entry: &entry{value: value, expiration: expiration},
		key:   key,
		cost:  cost,
	}

	//c.putEntryChannel <- entry
	//return true

	select {
	case c.putEntryChannel <- entry:
		return true
	default:
		return false
	}
}

func (c *Cache) PutSync(key interface{}, value interface{}, cost int64) bool {
	return c.PutWithTTLSync(key, value, cost, 0)
}

func (c *Cache) PutWithTTLSync(key interface{}, value interface{}, cost int64, ttl time.Duration) bool {
	if c == nil || key == nil || ttl < 0 || cost < 0 {
		return false
	}

	var expiration time.Time
	switch {
	case ttl == 0:
		break
	case ttl < 0:
		return false
	default:
		expiration = time.Now().Add(ttl)
	}

	entry := &bigEntry{
		entry: &entry{value: value, expiration: expiration},
		key:   key,
		cost:  cost,
	}

	c._save(entry)
	return true
}
