package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"

	"context"
	"net/http"
	"slices"
)

func (s *ModuleService) Update(ctx context.Context, module moduleDomain.Module, userID int) (
	*moduleDomain.Module, error) {

	if module.ID <= 0 {
		return nil, coreHttp.NewValidationError("module id", "module id is required")
	}

	if module.OwnerId != userID {
		return nil, coreHttp.ErrPermissionDenied
	}
	if module.Name == "" {
		return nil, coreHttp.NewValidationError("name", "name is required")
	}

	for _, webhook := range module.Webhooks {
		if webhook.ModuleID != module.ID {
			return nil, coreHttp.ErrInvalidBody
		}
		if webhook.ID <= 0 {
			return nil, coreHttp.ErrInvalidBody
		}
		methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
		if !slices.Contains(methods, webhook.Method) {
			return nil, coreHttp.ErrInvalidBody
		}
	}

	updatedModule, err := s.moduleRepo.UpdateModule(ctx, module, userID)
	if err != nil {
		return nil, err
	}

	s.toKafka("update", userID, updatedModule.Webhooks)

	return updatedModule, nil
}
