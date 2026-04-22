package mock

import (
	"auto-http-fetcher/internal/analytics/domain"
	"context"
)

type MockAnalyticsRepository struct {
	GetFunc func(ctx context.Context) (*domain.Analytics, error)
}

func (m *MockAnalyticsRepository) Get(ctx context.Context) (*domain.Analytics, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx)
	}
	return nil, nil
}
