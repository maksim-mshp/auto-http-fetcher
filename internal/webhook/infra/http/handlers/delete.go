package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	"errors"
	"net/http"
	"strconv"
)

// Delete godoc
// @Summary		Удалить вебхук
// @Description	Удаляет вебхук из модуля текущего пользователя.
// @Tags		Вебхуки
// @Produce		json
// @Param		module_id path int true "ID модуля"
// @Param		webhook_id path int true "ID вебхука"
// @Success		204
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		409 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{module_id}/webhook/{webhook_id} [delete]
// @Security	Bearer
func (wh *WebhookHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	wh.logger.Debug("webhook delete endpoint called")
	user, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrUnauthorized)
		return
	}
	module := r.PathValue("module_id")
	moduleInt, err := strconv.Atoi(module)
	if err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInvalidModuleID)
		return
	}
	webhookId := r.PathValue("webhook_id")
	webhookIdInt, err := strconv.Atoi(webhookId)
	if err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInvalidWebhookID)
		return
	}

	err = wh.moduleService.Delete(r.Context(), webhookIdInt, moduleInt, user.ID)
	if err != nil {
		var errAPI coreHttp.APIError
		if errors.As(err, &errAPI) {
			coreHttp.SendErrorJSON(wh.logger, w, &errAPI)
			return
		}
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
