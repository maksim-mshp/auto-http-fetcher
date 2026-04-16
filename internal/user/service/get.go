package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/security"
	"auto-http-fetcher/internal/user/domain"
	"context"
)

func (u *UserService) Get(ctx context.Context, checkUser *domain.User) (string, error) {

	user, err := u.userRepo.GetByEmail(ctx, checkUser.Email.String())
	if err != nil {
		u.logger.Error("get user by email error", "error", err)
		return "", err
	}

	if ok := security.Verify(user.Password, checkUser.Password); !ok {
		u.logger.Error("verify user password error", "user email", user.Email.String())
		return "", coreHttp.ErrVerificationFailed
	}

	token, err := u.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		u.logger.Error("generate access token error", "error", err)
		return "", coreHttp.ErrInternal
	}

	u.logger.Info("get user success", "user email", user.Email.String())
	return token, nil
}
