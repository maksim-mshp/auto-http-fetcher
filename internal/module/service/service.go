package service

import (
	"auto-http-fetcher/internal/module/domain"
	"context"
)

type ModuleService interface {
	CreateModule(context.Context, domain.Module) (*domain.Module, error)
	UpdateModule(context.Context, domain.Module) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID string) error
	GetModule(ctx context.Context, moduleID string) (*domain.Module, error)
	GetModuleList(context.Context) ([]*domain.Module, error)
}
