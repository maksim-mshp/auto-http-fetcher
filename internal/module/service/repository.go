package service

import (
	"auto-http-fetcher/internal/module/domain"
	"context"
)

type ModuleRepository interface {
	CreateModule(ctx context.Context, module domain.Module) (*domain.Module, error)
	UpdateModule(ctx context.Context, module domain.Module) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID int) error
	GetModule(ctx context.Context, moduleID int) (*domain.Module, error)
	GetModuleList(ctx context.Context) ([]*domain.Module, error)

	NewModuleID(ctx context.Context) (int, error)
}
