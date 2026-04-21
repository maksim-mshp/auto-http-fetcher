package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleDomain "auto-http-fetcher/internal/module/domain"

	"context"
	"net/http"
	"slices"
)

func (s *ModuleService) Create(ctx context.Context, module moduleDomain.Module, userID int) (
	*moduleDomain.Module, error) {
	if userID <= 0 {
		return nil, coreHttp.NewValidationError("user id", "userID is required")
	}

	if module.Name == "" {
		return nil, coreHttp.NewValidationError("name", "name is required")
	}

	for _, webhook := range module.Webhooks {
		methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
		if !slices.Contains(methods, webhook.Method) {
			return nil, coreHttp.ErrInvalidBody
		}
	}

	newModule, err := s.moduleRepo.CreateModule(ctx, module, userID)
	if err != nil {
		return nil, err
	}

	s.toKafka("create", userID, newModule.Webhooks)

	return newModule, nil
}
