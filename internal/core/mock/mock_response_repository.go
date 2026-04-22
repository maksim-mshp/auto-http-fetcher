package mock

import (
	"auto-http-fetcher/internal/response/domain"
	"context"
)

type MockResponseRepository struct {
	SaveFunc            func(ctx context.Context, r *domain.Response) error
	FindByIDFunc        func(ctx context.Context, id string) (*domain.Response, error)
	FindByWebhookIDFunc func(ctx context.Context, webhookID string) ([]*domain.Response, error)
}

func (m *MockResponseRepository) Save(ctx context.Context, r *domain.Response) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, r)
	}
	return nil
}

func (m *MockResponseRepository) FindByID(ctx context.Context, id string) (*domain.Response, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockResponseRepository) FindByWebhookID(ctx context.Context, webhookID string) ([]*domain.Response, error) {
	if m.FindByWebhookIDFunc != nil {
		return m.FindByWebhookIDFunc(ctx, webhookID)
	}
	return nil, nil
}
