package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	moduleHttp "auto-http-fetcher/internal/module/infra/http"

	"errors"
	"net/http"
)

// Update godoc
// @Summary		Обновить модуль
// @Description	Полностью обновляет данные модуля. Модуль должен принадлежать текущему пользователю.
// @Tags		Модули
// @Accept		json
// @Produce		json
// @Param		request body ModuleRequestResponse true "Данные модуля"
// @Success		200 {object} ModuleRequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		403 {object} APIError
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/ [put]
// @Security	Bearer
func (m *ModuleHandlers) Update(w http.ResponseWriter, r *http.Request) {
	m.logger.Debug("module update endpoint called")

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

	newModule, err := m.moduleService.Update(r.Context(), *domainModule, user.ID)
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
	}, http.StatusOK)
}
