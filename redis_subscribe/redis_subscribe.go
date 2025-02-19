package redis_subscribe

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

type Subscribe struct {
	redis *redis.Client
}

func NewSubscribe(redis *redis.Client) *Subscribe {
	return &Subscribe{redis: redis}
}

// Sub 订阅消息
// TODO 未验证 redis 断线重连，不确定有没有问题
func (s *Subscribe) Sub(ctx context.Context, topic string, r chan string) error {
	pubsub := s.redis.Subscribe(ctx, topic)
	err := pubsub.Subscribe(ctx, topic)
	if err != nil {
		return fmt.Errorf("redis Subscribe error: %w", err)
	}

	ch := pubsub.Channel()

	log.Printf("redis Subscribe stating")
	defer func() {
		log.Printf("redis Subscribe done")
	}()

	for {
		select {
		case s, ok := <-ch:
			if ok {
				r <- s.Payload
			} else {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Subscribe) Pub(ctx context.Context, topic string, payload string) error {
	_, err := s.redis.Publish(ctx, topic, payload).Result()
	if err != nil {
		return fmt.Errorf("redis Publish error: %w", err)
	}

	return nil
}
