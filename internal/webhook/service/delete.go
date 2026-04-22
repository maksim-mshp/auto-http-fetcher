package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/webhook/domain"
	"context"
	"errors"
)

func (s *WebhookService) Delete(ctx context.Context, webhookID, moduleID, userID int) error {
	if userID <= 0 {
		return coreHttp.NewValidationError("user id", "userID is required")
	}

	if moduleID <= 0 {
		return coreHttp.NewValidationError("module id", "moduleID is required")
	}

	if webhookID <= 0 {
		return coreHttp.NewValidationError("webhook id", "webhookID is required")
	}

	err := s.webhookRepo.DeleteWebhook(ctx, webhookID, moduleID, userID)
	if err != nil {
		if errors.As(err, &coreHttp.ErrWebhookNotFound) {
			return nil
		}
		return err
	}

	s.toKafka("delete", userID, &domain.Webhook{
		ID:       webhookID,
		ModuleID: moduleID,
	})

	return nil
}
