package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	webhookHttp "auto-http-fetcher/internal/webhook/infra/http"
	"auto-http-fetcher/internal/webhook/service"
	"log/slog"
)

type APIError = coreHttp.APIError
type WebhookDTORequestResponse = webhookHttp.WebhookDTORequestResponse

type WebhookHandlers struct {
	moduleService service.WebhookService
	logger        *slog.Logger
}

func NewWebhookHandlers(logger *slog.Logger, webhookService service.WebhookService) *WebhookHandlers {
	return &WebhookHandlers{webhookService, logger}
}
