package service

import (
	"auto-http-fetcher/internal/analytics/domain"
	"auto-http-fetcher/internal/core/mock"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsService_Get_FromCache(t *testing.T) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	expectedAnalytics := &domain.Analytics{
		TotalCalls:   100,
		SuccessCalls: 80,
		FailedCalls:  20,
	}
	data, _ := json.Marshal(expectedAnalytics)
	_ = mr.Set("analytics", string(data))

	mockRepo := &mock.MockAnalyticsRepository{
		GetFunc: func(ctx context.Context) (*domain.Analytics, error) {
			t.Fatal("Repository should not be called")
			return nil, nil
		},
	}

	service := NewAnalyticsService(mockRepo, redisClient, 60*time.Second)

	result, err := service.Get(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 100, result.TotalCalls)
	assert.Equal(t, 80, result.SuccessCalls)
	assert.Equal(t, 20, result.FailedCalls)
}

func TestAnalyticsService_Get_FromRepository(t *testing.T) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	expectedAnalytics := &domain.Analytics{
		TotalCalls:   200,
		SuccessCalls: 150,
		FailedCalls:  50,
	}

	mockRepo := &mock.MockAnalyticsRepository{
		GetFunc: func(ctx context.Context) (*domain.Analytics, error) {
			return expectedAnalytics, nil
		},
	}

	service := NewAnalyticsService(mockRepo, redisClient, 60*time.Second)

	result, err := service.Get(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, result.TotalCalls)

	stored, _ := mr.Get("analytics")
	var cached domain.Analytics
	_ = json.Unmarshal([]byte(stored), &cached)
	assert.Equal(t, 200, cached.TotalCalls)
}

func TestAnalyticsService_Get_RepositoryError(t *testing.T) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockRepo := &mock.MockAnalyticsRepository{
		GetFunc: func(ctx context.Context) (*domain.Analytics, error) {
			return nil, assert.AnError
		},
	}

	service := NewAnalyticsService(mockRepo, redisClient, 60*time.Second)

	result, err := service.Get(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAnalyticsService_Get_EmptyCache(t *testing.T) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	expectedAnalytics := &domain.Analytics{
		TotalCalls:   0,
		SuccessCalls: 0,
		FailedCalls:  0,
	}

	mockRepo := &mock.MockAnalyticsRepository{
		GetFunc: func(ctx context.Context) (*domain.Analytics, error) {
			return expectedAnalytics, nil
		},
	}

	service := NewAnalyticsService(mockRepo, redisClient, 60*time.Second)

	result, err := service.Get(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.TotalCalls)
}
