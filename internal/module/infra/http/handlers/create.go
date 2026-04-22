package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	moduleHttp "auto-http-fetcher/internal/module/infra/http"

	"errors"
	"net/http"
)

func (m *ModuleHandlers) Create(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrUnauthorized)
		return
	}

	var req moduleHttp.ModuleRequestResponse
	if err := coreHttp.ParseJSONBody(m.logger, r, &req); err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	domainModule, err := req.Module.ToDomain()
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	newModule, err := m.moduleService.Create(r.Context(), *domainModule, user.ID)
	if err != nil {
		var errAPI coreHttp.APIError
		if errors.As(err, &errAPI) {
			coreHttp.SendErrorJSON(m.logger, w, &errAPI)
			return
		}
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInternal)
		return
	}

	coreHttp.SendJSON(m.logger, w, &moduleHttp.ModuleRequestResponse{
		Module: *moduleHttp.ModuleToDTO(newModule),
	}, http.StatusCreated)
}
