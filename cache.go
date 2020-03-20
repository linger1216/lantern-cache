package lantern

import (
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
	policy  *defaultPolicy
	store   *storeShared
	hasher  hasher
	onEvict onEvictFunc
}

type Config struct {
	Shards  uint64
	MaxCost uint64
	OnEvict onEvictFunc
	Hash    string
}

func NewLanternCache(conf *Config) *Cache {
	ret := &Cache{}
	ret.policy = newDefaultPolicy(10000, conf.MaxCost)
	// todo
	// 不同引擎config
	ret.store = newStoreShared(conf.Shards)

	switch strings.ToLower(conf.Hash) {
	case "fnvxx":
		fallthrough
	default:
		ret.hasher = newFnvXX()
	}

	ret.onEvict = conf.OnEvict
	return ret
}

func (c *Cache) SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) error {

	if c == nil || key == nil {
		return ErrorInvalidPara
	}

	if ttl <= 0 {
		return ErrorNoExpiration
	}

	hash, conflict := c.hasher.hash(key)
	entry := &entry{
		key:        hash,
		conflict:   conflict,
		value:      value,
		cost:       cost,
		expiration: time.Now().Add(ttl),
	}

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

	for i := range evicts {
		// todo 删除为什么要用0
		evicts[i].conflict, evicts[i].value = c.store.Del(evicts[i].key, 0)
		if c.onEvict != nil {
			c.onEvict(evicts[i].key, evicts[i].conflict, evicts[i].value, evicts[i].cost)
		}
	}
	return nil
}
