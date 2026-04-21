package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"

	"context"
)

func (s *ModuleService) Delete(ctx context.Context, moduleID, userID int) error {
	if moduleID <= 0 {
		return coreHttp.NewValidationError("module id", "moduleID is required")
	}
	if userID <= 0 {
		return coreHttp.NewValidationError("user id", "userID is required")
	}

	if err := s.moduleRepo.DeleteModule(ctx, moduleID, userID); err != nil {
		return err
	}

	// TODO: kafka

	return nil
}
