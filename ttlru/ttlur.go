package memcache

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"time"
)

type TTLru[T any] struct {
	cache *lru.Cache[string, memCacheItem[T]]
}

func NewTTLru[T any](size int) *TTLru[T] {
	c, err := lru.New[string, memCacheItem[T]](size)
	if err != nil {
		panic(err)
	}

	go func() {
		for range time.Tick(4 * time.Hour) {
			for _, k := range c.Keys() {
				v, ok := c.Get(k)
				if !ok {
					continue
				}
				if v.expireAt.Before(time.Now()) {
					c.Remove(k)
				}
			}
		}
	}()

	return &TTLru[T]{
		cache: c,
	}
}

type memCacheItem[T any] struct {
	i        T
	expireAt time.Time
}

func (c *TTLru[T]) Get(key string) (t T, exist bool) {
	x, ok := c.cache.Get(key)
	if !ok {
		return
	}
	if x.expireAt.Before(time.Now()) {
		return
	}
	return x.i, true
}

func (c *TTLru[T]) Set(key string, v T, ttl time.Duration) {
	c.cache.Add(key, memCacheItem[T]{
		i:        v,
		expireAt: time.Now().Add(ttl),
	})
}
