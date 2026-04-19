package mock

import (
	"context"

	"auto-http-fetcher/internal/user/domain"
)

type MockUserRepository struct {
	CreateFunc     func(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	GetByIDFunc    func(ctx context.Context, id int) (*domain.User, error)
	UpdateFunc     func(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteFunc     func(ctx context.Context, id int) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
