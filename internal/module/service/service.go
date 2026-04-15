package service

import "log/slog"

type ModuleService struct {
	logger *slog.Logger
	repo   ModuleRepository
}

func NewModuleService(logger *slog.Logger, repo ModuleRepository) *ModuleService {
	return &ModuleService{logger, repo}
}
