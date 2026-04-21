package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	moduleHttp "auto-http-fetcher/internal/module/infra/http"

	"errors"
	"net/http"
)

func (m *ModuleHandlers) GetList(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrUnauthorized)
		return
	}

	modules, err := m.moduleService.List(r.Context(), user.ID)
	if err != nil {
		var errAPI coreHttp.APIError
		if errors.As(err, &errAPI) {
			coreHttp.SendErrorJSON(m.logger, w, &errAPI)
			return
		}
		coreHttp.SendErrorJSON(m.logger, w, &coreHttp.ErrInternal)
		return
	}

	modulesDTO := make([]*moduleHttp.ModuleDTO, len(modules))
	for i, module := range modules {
		modulesDTO[i] = moduleHttp.ModuleToDTO(module)
	}

	coreHttp.SendJSON(m.logger, w, moduleHttp.ModuleList{
		Modules: modulesDTO,
	}, http.StatusOK)
}
