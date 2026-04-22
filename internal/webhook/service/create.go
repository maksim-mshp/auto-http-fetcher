package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/webhook/domain"
	"context"
	"net/http"
	"slices"
)

func (s *WebhookService) Create(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (
	*domain.Webhook, error) {
	if userID <= 0 {
		return nil, coreHttp.NewValidationError("user id", "userID is required")
	}

	if moduleID <= 0 {
		return nil, coreHttp.NewValidationError("module id", "moduleID is required")
	}

	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
	if !slices.Contains(methods, webhook.Method) {
		return nil, coreHttp.ErrInvalidBody
	}

	newWebhook, err := s.webhookRepo.CreateWebhook(ctx, webhook, moduleID, userID)
	if err != nil {
		return nil, err
	}

	s.toKafka("create", userID, newWebhook)

	return newWebhook, nil
}
