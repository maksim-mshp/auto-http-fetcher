package mock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type MockRedisClient struct {
	GetFunc func(ctx context.Context, key string) *redis.StringCmd
	SetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return redis.NewStringResult("", redis.Nil)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, expiration)
	}
	return redis.NewStatusResult("OK", nil)
}
