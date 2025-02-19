package ttlru

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"time"
)

type TTLru[T any] struct {
	cache   *lru.Cache[string, memCacheItem[T]]
	maxSize int
}

func NewTTLru[T any](size int) *TTLru[T] {
	c, err := lru.New[string, memCacheItem[T]](size)
	if err != nil {
		panic(err)
	}
	return &TTLru[T]{
		cache:   c,
		maxSize: size,
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

func (c *TTLru[T]) Size() (curr, max int) {
	return c.cache.Len(), c.maxSize
}
func (c *TTLru[T]) Keys() []string {
	return c.cache.Keys()
}

func (c *TTLru[T]) Set(key string, v T, ttl time.Duration) {
	c.cache.Add(key, memCacheItem[T]{
		i:        v,
		expireAt: time.Now().Add(ttl),
	})
}

func (c *TTLru[T]) Delete(key string) {
	c.cache.Remove(key)
}
