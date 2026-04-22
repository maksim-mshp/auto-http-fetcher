package service

import (
	"auto-http-fetcher/internal/webhook/domain"
	"context"
)

type WebhookRepository interface {
	CreateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID, moduleID, userID int) error
}
