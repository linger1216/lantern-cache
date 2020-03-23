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
	policy           *defaultPolicy
	store            *storeShared
	hasher           hasher
	accessRingBuffer *ringPoll
	cleanupTicker    *time.Ticker
	stopChannel      chan struct{}
	putEntryChannel  chan *entry
	onEvict          onEvictFunc
}

type Config struct {
	Shards              uint64
	MaxKeyCount         uint64
	BucketInterval      int64
	MaxCost             uint64
	MaxAccessRingBuffer uint64
	PutEntryBuffer      uint64
	OnEvict             onEvictFunc
	Hash                string
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

	if conf.MaxAccessRingBuffer == 0 {
		conf.MaxAccessRingBuffer = 64
	}

	ret.accessRingBuffer = newRingPool(ret.policy, conf.MaxAccessRingBuffer)

	ret.onEvict = conf.OnEvict
	ret.cleanupTicker = time.NewTicker(time.Duration(conf.BucketInterval) * time.Second / 2)
	ret.stopChannel = make(chan struct{})

	if conf.PutEntryBuffer == 0 {
		conf.PutEntryBuffer = 32 * 1024
	}
	ret.putEntryChannel = make(chan *entry, conf.PutEntryBuffer)
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
			evicts, saved, err := c.policy.put(entry.key, entry.cost)
			if err != nil {
				fmt.Printf("err:%s\n", err.Error())
				break
			}

			if saved {
				err = c.store.Put(entry)
				if err != nil {
					fmt.Printf("err:%s\n", err.Error())
					break
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
