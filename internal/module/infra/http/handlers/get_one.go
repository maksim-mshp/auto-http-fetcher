package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	moduleHttp "auto-http-fetcher/internal/module/infra/http"

	"errors"
	"net/http"
	"strconv"
)

func (m *ModuleHandlers) GetOne(w http.ResponseWriter, r *http.Request) {
	m.logger.Debug("module get one endpoint called")

	user, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrUnauthorized)
		return
	}

	moduleGetRequest := r.PathValue("id")
	moduleID, err := strconv.Atoi(moduleGetRequest)
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	module, err := m.moduleService.Get(r.Context(), moduleID, user.ID)
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
		Module: *moduleHttp.ModuleToDTO(module),
	}, http.StatusOK)
}
