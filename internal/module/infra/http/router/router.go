package router

import (
	"auto-http-fetcher/internal/core/middleware"
	"auto-http-fetcher/internal/core/security"
	modulesHandlers "auto-http-fetcher/internal/module/infra/http/handlers"
	"log/slog"
	"net/http"
)

func GetModulesRouter(logger *slog.Logger, jwt *security.JWT, moduleHandles *modulesHandlers.ModuleHandlers) http.Handler {
	module := http.NewServeMux()
	module.HandleFunc("POST /module/", moduleHandles.Create)
	module.HandleFunc("PUT /module/", moduleHandles.Update)
	module.HandleFunc("DELETE /module/{id}", moduleHandles.Delete)
	module.HandleFunc("GET /module/{id}", moduleHandles.GetOne)
	module.HandleFunc("GET /modules/", moduleHandles.GetList)

	return middleware.Logger(logger)(
		middleware.AuthMiddleware(jwt)(module))
}
