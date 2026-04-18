package service

import (
	"auto-http-fetcher/internal/core/security"
	"log/slog"
)

type UserService struct {
	logger   *slog.Logger
	jwt      *security.JWT
	userRepo UserRepository
}

func NewUserService(logger *slog.Logger, jwt *security.JWT, userRepo UserRepository) *UserService {
	return &UserService{logger: logger, jwt: jwt, userRepo: userRepo}
}
