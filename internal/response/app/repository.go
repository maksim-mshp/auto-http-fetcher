package app

import (
	"auto-http-fetcher/internal/response/domain"
	"context"
)

type Repository interface {
	Save(ctx context.Context, r *domain.Response) error
	FindByID(ctx context.Context, id string) (*domain.Response, error)
	FindByWebhookID(ctx context.Context, webhookID string) ([]*domain.Response, error)
}
