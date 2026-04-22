package service

import (
	"auto-http-fetcher/internal/module/domain"

	"context"
)

type ModuleRepository interface {
	CreateModule(ctx context.Context, module domain.Module, userID int) (*domain.Module, error)
	UpdateModule(ctx context.Context, module domain.Module, userID int) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID, userID int) error
	GetModule(ctx context.Context, moduleID, userID int) (*domain.Module, error)
	GetModuleList(ctx context.Context, userID int) ([]*domain.Module, error)
}
