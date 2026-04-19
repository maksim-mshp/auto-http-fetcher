package service

import (
	"auto-http-fetcher/internal/analytics/domain"
	"context"
	"encoding/json"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type AnalyticsService struct {
	repo  Repository
	redis *redis.Client
	ttl   time.Duration
}

func NewAnalyticsService(repo Repository, redis *redis.Client, ttl time.Duration) *AnalyticsService {
	return &AnalyticsService{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (a *AnalyticsService) Get(ctx context.Context) (*domain.Analytics, error) {
	val, err := a.redis.Get(ctx, "analytics").Bytes()
	if err == nil {
		var analytics domain.Analytics
		if err = json.Unmarshal(val, &analytics); err != nil {
			return nil, err
		}
		return &analytics, nil
	}
	analytics, err := a.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(analytics)
	if err != nil {
		return nil, err
	}
	if err := a.redis.Set(ctx, "analytics", data, a.ttl).Err(); err != nil {
		return nil, err
	}
	return analytics, nil
}
