package handlers

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	moduleHttp "auto-http-fetcher/internal/module/infra/http"
	"auto-http-fetcher/internal/module/service"

	"log/slog"
)

type APIError = coreHttp.APIError
type ModuleRequestResponse = moduleHttp.ModuleRequestResponse
type ModuleList = moduleHttp.ModuleList

type ModuleHandlers struct {
	moduleService service.ModuleService
	logger        *slog.Logger
}

func NewModuleHandlers(logger *slog.Logger, moduleService service.ModuleService) *ModuleHandlers {
	return &ModuleHandlers{moduleService, logger}
}
