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

type onEvictFunc func(key uint64, conflict uint64, value interface{}, cost int64)

type Cache struct {
	policy        *defaultPolicy
	store         *storeShared
	hasher        hasher
	access        *ringPoll
	cleanupTicker *time.Ticker
	onEvict       onEvictFunc
}

type Config struct {
	Shards         uint64
	MaxKeyCount    uint64
	BucketInterval int64
	MaxCost        uint64
	MaxRingBuffer  uint64
	OnEvict        onEvictFunc
	Hash           string
}

func NewLanternCache(conf *Config) *Cache {
	ret := &Cache{}

	// todo
	// need more test for this value
	conf.BucketInterval = 5
	switch strings.ToLower(conf.Hash) {
	case "fnvxx":
		fallthrough
	default:
		ret.hasher = newFnvXX()
	}

	ret.policy = newDefaultPolicy(conf.MaxKeyCount, conf.MaxCost)

	// todo
	// 不同引擎config
	ret.store = newStoreShared(conf.Shards, conf.BucketInterval, conf.OnEvict)

	if conf.MaxRingBuffer == 0 {
		conf.MaxRingBuffer = 64
	}
	ret.access = newRingPool(ret.policy, conf.MaxRingBuffer)

	ret.onEvict = conf.OnEvict
	ret.cleanupTicker = time.NewTicker(time.Duration(conf.BucketInterval) * time.Second / 2)

	go ret.process()
	return ret
}

func (c *Cache) process() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.store.Clean()
		}
	}
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

	fmt.Printf("[put] key:%d conflict:%d val:%s cost:%d ttl:%fs\n",
		keyHash, conflict, value.(string), cost, ttl.Seconds())

	// evicts only has key and cost
	evicts, saved, err := c.policy.put(entry.key, entry.cost)
	if err != nil {
		return err
	}

	if saved {
		err = c.store.Put(entry)
		if err != nil {
			return err
		}
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
	return nil
}
