package handlers

import (
	"auto-http-fetcher/internal/module/service"

	"log/slog"
)

type ModuleHandlers struct {
	moduleService service.ModuleService
	logger        *slog.Logger
}

func NewModuleHandlers(logger *slog.Logger, moduleService service.ModuleService) *ModuleHandlers {
	return &ModuleHandlers{moduleService, logger}
}
