package router

import (
	"auto-http-fetcher/internal/core/middleware"
	"auto-http-fetcher/internal/core/security"
	modulesHandlers "auto-http-fetcher/internal/module/infra/http/handlers"
	webhookHandlers "auto-http-fetcher/internal/webhook/infra/http/handlers"
	"log/slog"
	"net/http"
)

func GetModulesRouter(logger *slog.Logger, jwt *security.JWT, moduleHandles *modulesHandlers.ModuleHandlers,
	webhookHandles *webhookHandlers.WebhookHandlers) http.Handler {
	module := http.NewServeMux()
	module.HandleFunc("POST /module/", moduleHandles.Create)
	module.HandleFunc("PUT /module/", moduleHandles.Update)
	module.HandleFunc("DELETE /module/{id}", moduleHandles.Delete)
	module.HandleFunc("GET /module/{id}", moduleHandles.GetOne)
	module.HandleFunc("GET /modules/", moduleHandles.GetList)

	module.HandleFunc("POST /module/{module_id}/webhook/", webhookHandles.Create)
	module.HandleFunc("PUT /module/{module_id}/webhook/", webhookHandles.Update)
	module.HandleFunc("DELETE /module/{module_id}/webhook/{webhook_id}", webhookHandles.Delete)

	return middleware.Logger(logger)(
		middleware.AuthMiddleware(jwt)(module))
}
