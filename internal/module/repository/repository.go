package repository

import (
	"auto-http-fetcher/internal/module/domain"
	"context"
)

type ModuleRepository interface {
	CreateModule(ctx context.Context, module domain.Module) (*domain.Module, error)
	UpdateModule(ctx context.Context, module domain.Module) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID string) error
	GetModule(ctx context.Context, moduleID string) (*domain.Module, error)
	GetModuleList(ctx context.Context) ([]*domain.Module, error)
}
