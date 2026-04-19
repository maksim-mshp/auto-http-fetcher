package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"
	"context"
)

func (s *ModuleService) List(ctx context.Context, userID int) (
	[]*moduleDomain.Module, error) {
	if userID <= 0 {
		return nil, coreHttp.NewValidationError("user id", "userID is required")
	}

	modules, err := s.moduleRepo.GetModuleList(ctx, userID)
	if err != nil {
		return nil, err
	}

	return modules, nil
}
