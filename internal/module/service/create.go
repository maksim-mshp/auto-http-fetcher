package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"
	"context"
)

func (s *ModuleService) Create(ctx context.Context, module moduleDomain.Module, userID int) (
	*moduleDomain.Module, error) {
	if userID <= 0 {
		return nil, coreHttp.NewValidationError("user id", "userID is required")
	}

	if module.Name == "" {
		return nil, coreHttp.NewValidationError("name", "name is required")
	}

	newModule, err := s.moduleRepo.CreateModule(ctx, module, userID)
	if err != nil {
		return nil, err
	}

	// TODO: kafka

	return newModule, nil
}
