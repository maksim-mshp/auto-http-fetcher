package mock

import (
	"auto-http-fetcher/internal/webhook/domain"

	"context"

	"github.com/stretchr/testify/mock"
)

type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) CreateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error) {
	args := m.Called(ctx, webhook, moduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) UpdateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error) {
	args := m.Called(ctx, webhook, moduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockWebhookRepository) DeleteWebhook(ctx context.Context, webhookID, moduleID, userID int) error {
	args := m.Called(ctx, webhookID, moduleID, userID)
	return args.Error(0)
}

func (m *MockWebhookRepository) GetWebhook(ctx context.Context, webhookID, moduleID, userID int) (*domain.Webhook, error) {
	args := m.Called(ctx, webhookID, moduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}
