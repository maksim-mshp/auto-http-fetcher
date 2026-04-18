package service

import (
	"auto-http-fetcher/internal/core/security"
	"auto-http-fetcher/internal/user/domain"
	"context"
)

func (u *UserService) Create(ctx context.Context, user *domain.User) (*domain.User, error) {

	if err := domain.ValidateUser(user); err != nil {
		return nil, err
	}

	hashedPassword, err := security.Hash(user.Password)
	if err != nil {
		return nil, err
	}

	createdUser, err := u.userRepo.Create(ctx, &domain.User{
		Email:    user.Email,
		Password: hashedPassword,
		Name:     user.Name,
	})
	if err != nil {
		return nil, err
	}
	createdUser.Password = ""

	return createdUser, nil
}
