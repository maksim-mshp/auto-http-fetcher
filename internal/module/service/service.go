package service

import "log/slog"

type ModuleService struct {
	logger     *slog.Logger
	moduleRepo ModuleRepository
}

func NewModuleService(logger *slog.Logger, repo ModuleRepository) *ModuleService {
	return &ModuleService{logger, repo}
}
