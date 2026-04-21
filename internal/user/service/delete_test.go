package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/core/mock"
	"auto-http-fetcher/internal/core/security"

	"github.com/stretchr/testify/assert"
)

func TestUserService_Delete_Success(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		DeleteFunc: func(ctx context.Context, id int) error {
			assert.Equal(t, 1, id)
			return nil
		},
	}

	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	err := service.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestUserService_Delete_InvalidUserID(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}
	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)
	err := service.Delete(context.Background(), -1)

	assert.Error(t, err)
	assert.Equal(t, coreHttp.ErrInvalidUserID, err)
}

func TestUserService_Delete_UserNotFound(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		DeleteFunc: func(ctx context.Context, id int) error {
			return coreHttp.ErrUserNotFound
		},
	}

	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	err := service.Delete(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, coreHttp.ErrUserNotFound, err)
}

func TestUserService_Delete_DatabaseError(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		DeleteFunc: func(ctx context.Context, id int) error {
			return errors.New("connection failed")
		},
	}
	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	err := service.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection failed")
}
