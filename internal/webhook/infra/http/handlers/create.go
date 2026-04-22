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

func (wh *WebhookHandlers) Create(w http.ResponseWriter, r *http.Request) {
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
