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

func TestUserService_Create_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		user          *domain.User
		mockError     error
		expectedError error
	}{
		{
			name: "success",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "John Doe",
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "invalid email",
			user: &domain.User{
				Email:    "invalid",
				Password: "password123",
				Name:     "John",
			},
			mockError:     nil,
			expectedError: coreHttp.NewValidationError("email", "email is invalid"),
		},
		{
			name: "weak password",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "123",
				Name:     "John",
			},
			mockError:     nil,
			expectedError: coreHttp.NewValidationError("password", "password is too short"),
		},
		{
			name: "duplicate user",
			user: &domain.User{
				Email:    "exists@example.com",
				Password: "password123",
				Name:     "John",
			},
			mockError:     coreHttp.ErrUserAlreadyExists,
			expectedError: coreHttp.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &domain.User{
						ID:    1,
						Email: user.Email,
						Name:  user.Name,
					}, nil
				},
			}
			logger := slog.Default()
			jwt := security.NewJWTService("123", 1*time.Hour)
			service := NewUserService(logger, jwt, mockRepo)
			result, err := service.Create(context.Background(), tt.user)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.user.Email, result.Email)
				assert.Equal(t, tt.user.Name, result.Name)
				assert.Empty(t, result.Password)
			}
		})
	}
}
