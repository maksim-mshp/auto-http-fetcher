package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	webhookHttp "auto-http-fetcher/internal/webhook/infra/http"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// Create godoc
// @Summary		Создать вебхук
// @Description	Добавляет вебхук в модуль текущего пользователя.
// @Tags		Вебхуки
// @Accept		json
// @Produce		json
// @Param		module_id path int true "ID модуля"
// @Param		request body WebhookDTORequestResponse true "Данные вебхука"
// @Success		201 {object} WebhookDTORequestResponse
// @Failure		400 {object} APIError
// @Failure		401 {string} string
// @Failure		404 {object} APIError
// @Failure		415 {object} APIError
// @Failure		500 {object} APIError
// @Router		/module/{module_id}/webhook/ [post]
// @Security	Bearer
func (wh *WebhookHandlers) Create(w http.ResponseWriter, r *http.Request) {
	wh.logger.Debug("webhook create endpoint called")

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

	var req webhookHttp.WebhookDTORequestResponse
	if err := coreHttp.ParseJSONBody(wh.logger, r, &req); err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	domainWebhook, err := req.Webhook.ToDomain()
	if err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInvalidBody)
		return
	}

	newWebhook, err := wh.moduleService.Create(r.Context(), *domainWebhook, moduleInt, user.ID)
	if err != nil {
		log.Println(err)
		var errAPI coreHttp.APIError
		if errors.As(err, &errAPI) {
			coreHttp.SendErrorJSON(wh.logger, w, &errAPI)
			return
		}
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInternal)
		return
	}

	coreHttp.SendJSON(wh.logger, w, &webhookHttp.WebhookDTORequestResponse{
		Webhook: webhookHttp.WebhookToDTO(newWebhook),
	}, http.StatusCreated)
}
