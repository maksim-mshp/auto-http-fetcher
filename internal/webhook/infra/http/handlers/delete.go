package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/middleware"
	webhookHttp "auto-http-fetcher/internal/webhook/infra/http"
	"errors"
	"net/http"
	"strconv"
)

func (wh *WebhookHandlers) Delete(w http.ResponseWriter, r *http.Request) {
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

	var req webhookHttp.WebhookDTORequestResponse
	if err := coreHttp.ParseJSONBody(wh.logger, r, &req); err != nil {
		coreHttp.SendErrorJSON(wh.logger, w, &coreHttp.ErrInvalidBody)
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
