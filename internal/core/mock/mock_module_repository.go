package mock

import (
	"auto-http-fetcher/internal/module/domain"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockModuleRepository struct {
	mock.Mock
}

func (m *MockModuleRepository) CreateModule(ctx context.Context, module domain.Module, userID int) (*domain.Module, error) {
	args := m.Called(ctx, module, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Module), args.Error(1)
}

func (m *MockModuleRepository) UpdateModule(ctx context.Context, module domain.Module, userID int) (*domain.Module, error) {
	args := m.Called(ctx, module, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Module), args.Error(1)
}

func (m *MockModuleRepository) DeleteModule(ctx context.Context, moduleID, userID int) error {
	args := m.Called(ctx, moduleID, userID)
	return args.Error(0)
}

func (m *MockModuleRepository) GetModule(ctx context.Context, moduleID, userID int) (*domain.Module, error) {
	args := m.Called(ctx, moduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Module), args.Error(1)
}

func (m *MockModuleRepository) GetModuleList(ctx context.Context, userID int) ([]*domain.Module, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Module), args.Error(1)
}
