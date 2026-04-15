package service

import (
	"auto-http-fetcher/internal/core/security"
	"auto-http-fetcher/internal/user/domain"
	"context"
	"fmt"
)

func (u *UserService) Get(ctx context.Context, checkUser *domain.User) (string, error) {

	user, err := u.userRepo.GetByEmail(ctx, checkUser.Email)
	if err != nil {
		return "", err
	}

	if ok := security.Verify(user.Password, checkUser.Password); !ok {
		return "", fmt.Errorf("incorrect password")
	}

	token, err := u.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}
