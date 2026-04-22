package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	testMock "auto-http-fetcher/internal/core/mock"
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"

	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestModuleService_Get(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name           string
		moduleID       int
		userID         int
		setupMocks     func(*testMock.MockModuleRepository)
		expectedModule *domainModule.Module
		expectedError  error
	}{
		{
			name:     "success - get module",
			moduleID: 10,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				expectedModule := &domainModule.Module{
					ID:          10,
					Name:        "Test Module",
					Description: "Test Description",
					OwnerId:     1,
					Webhooks: []*domainWebhook.Webhook{
						{
							ID:          1,
							Description: "Test webhook",
							Interval:    10 * time.Second,
							Timeout:     5 * time.Second,
							URL:         url.URL{Scheme: "http", Host: "example.com"},
							Method:      http.MethodPost,
							Headers:     http.Header{"Content-Type": []string{"application/json"}},
							Body:        []byte(`{"test": true}`),
						},
					},
				}
				repo.On("GetModule", mock.Anything, 10, 1).Return(expectedModule, nil)
			},
			expectedModule: &domainModule.Module{
				ID:          10,
				Name:        "Test Module",
				Description: "Test Description",
				OwnerId:     1,
			},
			expectedError: nil,
		},
		{
			name:     "fail - invalid module id (zero)",
			moduleID: 0,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModule: nil,
			expectedError:  coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:     "fail - invalid module id (negative)",
			moduleID: -5,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModule: nil,
			expectedError:  coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:     "fail - invalid user id (zero)",
			moduleID: 10,
			userID:   0,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModule: nil,
			expectedError:  coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:     "fail - invalid user id (negative)",
			moduleID: 10,
			userID:   -3,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModule: nil,
			expectedError:  coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:     "fail - module not found",
			moduleID: 999,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModule", mock.Anything, 999, 1).Return(nil, coreHttp.ErrModuleNotFound)
			},
			expectedModule: nil,
			expectedError:  coreHttp.ErrModuleNotFound,
		},
		{
			name:     "fail - permission denied",
			moduleID: 10,
			userID:   2,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModule", mock.Anything, 10, 2).Return(nil, coreHttp.ErrPermissionDenied)
			},
			expectedModule: nil,
			expectedError:  coreHttp.ErrPermissionDenied,
		},
		{
			name:     "fail - database error",
			moduleID: 10,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModule", mock.Anything, 10, 1).Return(nil, errors.New("database connection failed"))
			},
			expectedModule: nil,
			expectedError:  errors.New("database connection failed"),
		},
		{
			name:     "success - get module without webhooks",
			moduleID: 20,
			userID:   3,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				expectedModule := &domainModule.Module{
					ID:          20,
					Name:        "Simple Module",
					Description: "No webhooks",
					OwnerId:     3,
					Webhooks:    []*domainWebhook.Webhook{},
				}
				repo.On("GetModule", mock.Anything, 20, 3).Return(expectedModule, nil)
			},
			expectedModule: &domainModule.Module{
				ID:          20,
				Name:        "Simple Module",
				Description: "No webhooks",
				OwnerId:     3,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testMock.MockModuleRepository)
			mockProducer := new(testMock.MockProducer)
			mockDLQ := new(testMock.MockDLQ)

			tt.setupMocks(mockRepo)

			service := &ModuleService{
				logger:     logger,
				moduleRepo: mockRepo,
				kafka:      mockProducer,
				dlq:        mockDLQ,
			}

			ctx := context.Background()
			result, err := service.Get(ctx, tt.moduleID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedModule.ID, result.ID)
				assert.Equal(t, tt.expectedModule.Name, result.Name)
				assert.Equal(t, tt.expectedModule.OwnerId, result.OwnerId)
			}

			mockRepo.AssertExpectations(t)

			mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
			mockDLQ.AssertNotCalled(t, "Push", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
