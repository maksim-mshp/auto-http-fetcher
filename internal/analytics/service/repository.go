package service

import (
	"auto-http-fetcher/internal/analytics/domain"
	"context"
)

type Repository interface {
	Get(ctx context.Context) (*domain.Analytics, error)
}
