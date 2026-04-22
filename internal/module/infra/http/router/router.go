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
	module.HandleFunc("POST /api/v1/module/", moduleHandles.Create)
	module.HandleFunc("PUT /api/v1/module/", moduleHandles.Update)
	module.HandleFunc("DELETE /api/v1/module/{id}", moduleHandles.Delete)
	module.HandleFunc("GET /api/v1/module/{id}", moduleHandles.GetOne)
	module.HandleFunc("GET /api/v1/modules/", moduleHandles.GetList)

	module.HandleFunc("POST /api/v1/module/{module_id}/webhook/", webhookHandles.Create)
	module.HandleFunc("PUT /api/v1/module/{module_id}/webhook/", webhookHandles.Update)
	module.HandleFunc("DELETE /api/v1/module/{module_id}/webhook/{webhook_id}", webhookHandles.Delete)

	return middleware.Logger(logger)(
		middleware.AuthMiddleware(jwt)(module))
}
