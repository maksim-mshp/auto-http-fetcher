package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"
	"context"
)

func (s *ModuleService) Update(ctx context.Context, module moduleDomain.Module, userID int) (
	*moduleDomain.Module, error) {

	if module.ID <= 0 {
		return nil, coreHttp.NewValidationError("module id", "module id is required")
	}

	if module.Name == "" {
		return nil, coreHttp.NewValidationError("name", "name is required")
	}

	updatedModule, err := s.moduleRepo.UpdateModule(ctx, module, userID)
	if err != nil {
		return nil, err
	}

	// TODO: kafka

	return updatedModule, nil
}
