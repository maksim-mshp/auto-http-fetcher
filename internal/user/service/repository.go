package service

import (
	"auto-http-fetcher/internal/user/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, userID int) error
	GetByID(ctx context.Context, userID int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}
