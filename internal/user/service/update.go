package service

import (
	"auto-http-fetcher/internal/core/security"
	"auto-http-fetcher/internal/user/domain"
	"context"
)

func (u *UserService) Update(ctx context.Context, user *domain.User) (*domain.User, error) {

	if err := domain.ValidateUser(user); err != nil {
		return nil, err
	}

	hashedPassword, err := security.Hash(user.Password)
	if err != nil {
		return nil, err
	}

	updatedUser, err := u.userRepo.Update(ctx, &domain.User{
		Email:    user.Email,
		Password: hashedPassword,
		Name:     user.Name,
	})
	if err != nil {
		return nil, err
	}
	updatedUser.Password = ""

	return updatedUser, nil
}
