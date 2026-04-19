package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"
	"context"
)

func (s *ModuleService) Get(ctx context.Context, moduleID, userID int) (
	*moduleDomain.Module, error) {
	if moduleID <= 0 {
		return nil, coreHttp.NewValidationError("module id", "moduleID is required")
	}
	if userID <= 0 {
		return nil, coreHttp.NewValidationError("user id", "userID is required")
	}
	module, err := s.moduleRepo.GetModule(ctx, moduleID, userID)
	if err != nil {
		return nil, err
	}
	return module, nil
}
