package handlers

import (
	"auto-http-fetcher/internal/webhook/service"
	"log/slog"
)

type WebhookHandlers struct {
	moduleService service.WebhookService
	logger        *slog.Logger
}

func NewWebhookHandlers(logger *slog.Logger, webhookService service.WebhookService) *WebhookHandlers {
	return &WebhookHandlers{webhookService, logger}
}
