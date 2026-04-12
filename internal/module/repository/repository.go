package repository

import (
	"auto-http-fetcher/internal/module/domain"
	"context"
)

type ModuleRepository interface {
	CreateModule(context.Context, domain.Module) (*domain.Module, error)
	UpdateModule(context.Context, domain.Module) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID string) error
	GetModule(ctx context.Context, moduleID string) (*domain.Module, error)
	GetModuleList(context.Context) ([]*domain.Module, error)
}
