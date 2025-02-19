package ttlru

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/zbysir/doraemon/redis_subscribe"
	"log"
	"os"
	"strings"
	"time"
)

// 分布式的 TTLru 缓存, 使用 redis 订阅来实现多个 pod 之间的通信
// 支持通过依赖来清除缓存，比如 一个数据依赖了 A 和 B，只要 A 或者 B 其中一个变动，就清除缓存。

type DataWithSignal[T any] struct {
	Data   T
	Signal map[string]struct{}
}

type DistributedTTLru[T any] struct {
	ttlru      *TTLru[DataWithSignal[T]]
	subscriber *redis_subscribe.Subscribe
	scope      string // redis scope
}

// NewDistributedTTLru 创建一个可以被分布式清理的 TTLru
// size: 做多缓存多少个条目
// scope: redis scope
// redis: redis 客户端
// 一定要记得调用 Start，否则数据不会被清理。
func NewDistributedTTLru[T any](size int, scope string, redis *redis.Client) *DistributedTTLru[T] {
	c := NewTTLru[DataWithSignal[T]](size)

	return &DistributedTTLru[T]{
		ttlru:      c,
		subscriber: redis_subscribe.NewSubscribe(redis),
		scope:      scope,
	}
}

// Start 开始监听清理事件，一定记得 Start，否则数据不会被清理。
func (c *DistributedTTLru[T]) Start(ctx context.Context) (err error) {
	r := make(chan string, 100)
	defer close(r)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case signal := <-r:
				signals := strings.Split(signal, "@@")
				keys := c.deleteBySignal(signals...)
				log.Printf("DistributedTTLru delete by signals: %v, aff: %v, hostname: %s", signals, len(keys), os.Getenv("HOSTNAME"))
			}
		}
	}()
	err = c.subscriber.Sub(ctx, c.scope+"disttl_signal", r)
	if err != nil {
		return err
	}

	return nil
}

func (c *DistributedTTLru[T]) Get(key string) (t T, exist bool) {
	r, exist := c.ttlru.Get(key)
	if !exist {
		return
	}

	return r.Data, true
}

// Set 当 signals 改变时缓存也会被清空。
func (c *DistributedTTLru[T]) Set(key string, v T, ttl time.Duration, signals []string) {
	signalMap := make(map[string]struct{})
	for _, i := range signals {
		signalMap[i] = struct{}{}

	}
	c.ttlru.Set(key, DataWithSignal[T]{
		Data:   v,
		Signal: signalMap,
	}, ttl)
}

// Size 返回 lru 大小
func (c *DistributedTTLru[T]) Size() (curr, max int) {
	return c.ttlru.Size()
}

// UpdateSignal 发送清空信号，清空包含 signal 的数据
func (c *DistributedTTLru[T]) UpdateSignal(signals ...string) error {
	err := c.subscriber.Pub(context.Background(), c.scope+"disttl_signal", strings.Join(signals, "@@"))
	if err != nil {
		return err
	}

	return nil
}

// 遍历数据，删除包含 signal 的数据
func (c *DistributedTTLru[T]) deleteBySignal(signals ...string) (keys []string) {
	// 考虑倒排索引来优化所有 key 遍历的问题
	for _, key := range c.ttlru.Keys() {
		r, exist := c.ttlru.Get(key)
		if !exist {
			continue
		}

		var has bool
		for _, signal := range signals {
			if _, ok := r.Signal[signal]; ok {
				has = true
				break
			}
		}
		if has {
			c.ttlru.Delete(key)
			keys = append(keys, key)
		}
	}

	return
}
