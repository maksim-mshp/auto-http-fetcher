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

func TestUserService_Update_Success(t *testing.T) {
	mockRepo := &mock.MockUserRepository{
		UpdateFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
			return &domain.User{
				ID:    1,
				Email: user.Email,
				Name:  user.Name,
			}, nil
		},
	}
	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "newpassword",
		Name:     "Updated Name",
	}

	updated, err := service.Update(context.Background(), user)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "test@example.com", updated.Email)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Empty(t, updated.Password)
}

func TestUserService_Update_ValidationError(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}
	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)
	user := &domain.User{
		ID:    1,
		Email: "invalid-email",
		Name:  "Test",
	}

	updated, err := service.Update(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestUserService_Update_UserNotFound(t *testing.T) {

	mockRepo := &mock.MockUserRepository{
		UpdateFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
			return nil, coreHttp.ErrUserNotFound
		},
	}
	logger := slog.Default()
	jwt := security.NewJWTService("123", 1*time.Second)
	service := NewUserService(logger, jwt, mockRepo)

	user := &domain.User{
		ID:       999,
		Email:    "test@example.com",
		Password: "Pablo_pablo228",
		Name:     "Test",
	}

	updated, err := service.Update(context.Background(), user)
	assert.Error(t, err)
	assert.Equal(t, coreHttp.ErrUserNotFound, err)
	assert.Nil(t, updated)
}
