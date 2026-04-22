package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"

	"errors"
	"net/http"
	"strconv"
)

func (m *ModuleHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	m.logger.Debug("module delete endpoint called")

	user, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrUnauthorized)
		return
	}

	moduleDeleteRequest := r.PathValue("id")
	moduleID, err := strconv.Atoi(moduleDeleteRequest)
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	err = m.moduleService.Delete(r.Context(), moduleID, user.ID)
	if err != nil {
		var errAPI coreHttp.APIError
		if errors.As(err, &errAPI) {
			coreHttp.SendErrorJSON(m.logger, w, &errAPI)
			return
		}
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
