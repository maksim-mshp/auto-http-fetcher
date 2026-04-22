package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"context"
	"errors"
)

func (s *ModuleService) Delete(ctx context.Context, moduleID, userID int) error {
	if moduleID <= 0 {
		return coreHttp.NewValidationError("module id", "moduleID is required")
	}
	if userID <= 0 {
		return coreHttp.NewValidationError("user id", "userID is required")
	}

	err := s.moduleRepo.DeleteModule(ctx, moduleID, userID)
	if err != nil {
		if errors.As(err, &coreHttp.ErrModuleNotFound) {
			return nil
		}
		return err
	}

	return nil
}
