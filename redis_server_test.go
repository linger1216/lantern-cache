package lantern_cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
)

func TestRedisServer(t *testing.T) {
	ca := NewLanternCache(&Config{
		BucketCount:  256,
		MaxCapacity:  1024 * 1024 * 100,
		InitCapacity: 1024 * 1024 * 50,
	})

	server := NewRedisServer(":6379", ca)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 300)

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	{
		ca.Reset()
		key := "key"
		val := "val"
		err := client.Set(key, val, 0).Err()
		assert.Nil(t, err)

		actual, err := client.Get(key).Result()
		assert.Nil(t, err)
		assert.Equal(t, val, actual)
	}

	{
		ca.Reset()
		key := "key"
		val := "val"
		err := client.Set(key, val, time.Second).Err()
		assert.Nil(t, err)

		actual, err := client.Get(key).Result()
		assert.Nil(t, err)
		assert.Equal(t, val, actual)

		time.Sleep(2 * time.Second)

		// 超时返回redis.nil
		actual, err = client.Get(key).Result()
		assert.Equal(t, redis.Nil, err)
		assert.Equal(t, "", actual)
	}

	{
		ca.Reset()
		kvs := make([]string, 0, 20)
		keys := make([]string, 0, 10)
		vals := make([]interface{}, 0, 10)
		for i := 0; i < 10; i++ {
			keys = append(keys, fmt.Sprintf("key-%d", i))
			vals = append(vals, fmt.Sprintf("val-%d", i))
			kvs = append(kvs, fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i))
		}
		err := client.MSet(kvs).Err()
		assert.Nil(t, err)
		actual, err := client.MGet(keys...).Result()
		assert.Equal(t, nil, err)
		assert.Equal(t, vals, actual)
	}

	{
		ca.Reset()
		bucket := "bucket"
		key := "key"
		val := "val"
		err := client.HSet(bucket, key, val).Err()
		assert.Nil(t, err)

		actual, err := client.HGet(bucket, key).Result()
		assert.Nil(t, err)
		assert.Equal(t, val, actual)
	}

	{
		ca.Reset()
		bucket := "bucket"
		kvs := make([]string, 0, 20)
		keys := make([]string, 0, 10)
		vals := make([]interface{}, 0, 10)
		for i := 0; i < 10; i++ {
			keys = append(keys, fmt.Sprintf("key-%d", i))
			vals = append(vals, fmt.Sprintf("val-%d", i))
			kvs = append(kvs, fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i))
		}
		err := client.HMSet(bucket, kvs).Err()
		assert.Nil(t, err)
		actual, err := client.HMGet(bucket, keys...).Result()
		assert.Equal(t, nil, err)
		assert.Equal(t, vals, actual)
	}

	{
		ca.Reset()
		key := "key"
		val := "val"
		err := client.Set(key, val, 0).Err()
		assert.Nil(t, err)

		actual, err := client.Get(key).Result()
		assert.Nil(t, err)
		assert.Equal(t, val, actual)

		client.Del(key)
		actual, err = client.Get(key).Result()
		assert.Equal(t, redis.Nil, err)
		assert.Equal(t, "", actual)
	}
}
