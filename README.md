# lantern-cache

idea from fastcache


- thread safe

### usage
```
	b := lantern_cache.NewLanternCache(&lantern_cache.Config{
		BucketCount:          512,
		MaxCapacity:          1024 * 1024 * 1024,
	})

	key1 := []byte("key1")
	val1 := []byte("val1")
	err := b.Put(key1, val1)
	if err != nil {
		panic(err)
	}

	actual, err := b.Get(key1)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(actual, val1) {
		panic("not equal")
	}
```


### benchmark
