package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/mock"
	"auto-http-fetcher/internal/core/security"
	"auto-http-fetcher/internal/user/domain"

	"github.com/stretchr/testify/assert"
)

func TestUserService_Get_WhenUserDoesNotExist(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, coreHttp.ErrUserNotFound
		},
	}

	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	token, err := service.Get(context.Background(), &domain.User{
		Email:    "new@example.com",
		Password: "pass123",
	})

	assert.Error(t, err)
	assert.Equal(t, coreHttp.ErrUserNotFound, err)
	assert.Empty(t, token)
}

func TestUserService_Get_WhenPasswordIsWrong(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:       1,
				Email:    email,
				Password: "hashed_password",
			}, nil
		},
	}

	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	token, err := service.Get(context.Background(), &domain.User{
		Email:    "test@example.com",
		Password: "wrong",
	})

	assert.Error(t, err)
	assert.Equal(t, coreHttp.ErrVerificationFailed, err)
	assert.Empty(t, token)
}
