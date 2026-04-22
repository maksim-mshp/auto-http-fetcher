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

func TestModuleService_Update(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		userID        int
		module        domainModule.Module
		setupMocks    func(*testMock.MockModuleRepository)
		expectedError error
	}{
		{
			name:   "success - update module",
			userID: 1,
			module: domainModule.Module{
				ID:      10,
				Name:    "Updated Module Name",
				OwnerId: 1,
			},
			setupMocks: func(repo *testMock.MockModuleRepository) {
				updatedModule := &domainModule.Module{
					ID:      10,
					Name:    "Updated Module Name",
					OwnerId: 1,
				}
				repo.On("UpdateModule", mock.Anything, mock.Anything, 1).Return(updatedModule, nil)
			},
			expectedError: nil,
		},
		{
			name:   "fail - invalid module id (zero)",
			userID: 1,
			module: domainModule.Module{
				ID:      0,
				Name:    "Test Module",
				OwnerId: 1,
			},
			setupMocks:    func(repo *testMock.MockModuleRepository) {},
			expectedError: coreHttp.NewValidationError("module id", "module id is required"),
		},
		{
			name:   "fail - invalid module id (negative)",
			userID: 1,
			module: domainModule.Module{
				ID:      -5,
				Name:    "Test Module",
				OwnerId: 1,
			},
			setupMocks:    func(repo *testMock.MockModuleRepository) {},
			expectedError: coreHttp.NewValidationError("module id", "module id is required"),
		},
		{
			name:   "fail - permission denied (owner_id mismatch)",
			userID: 1,
			module: domainModule.Module{
				ID:      10,
				Name:    "Test Module",
				OwnerId: 2,
			},
			setupMocks:    func(repo *testMock.MockModuleRepository) {},
			expectedError: coreHttp.ErrPermissionDenied,
		},
		{
			name:   "fail - empty name",
			userID: 1,
			module: domainModule.Module{
				ID:      10,
				Name:    "",
				OwnerId: 1,
			},
			setupMocks:    func(repo *testMock.MockModuleRepository) {},
			expectedError: coreHttp.NewValidationError("name", "name is required"),
		},
		{
			name:   "fail - repository error",
			userID: 1,
			module: domainModule.Module{
				ID:      10,
				Name:    "Test Module",
				OwnerId: 1,
			},
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("UpdateModule", mock.Anything, mock.Anything, 1).Return(
					nil, errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:   "success - update with all fields",
			userID: 1,
			module: domainModule.Module{
				ID:          10,
				Name:        "Full Module Update",
				Description: "New description",
				OwnerId:     1,
				Webhooks: []*domainWebhook.Webhook{
					{
						ID:          1,
						Description: "Updated webhook",
						Interval:    15 * time.Second,
						Timeout:     10 * time.Second,
						URL:         url.URL{Scheme: "https", Host: "newapi.com"},
						Method:      http.MethodPut,
						Headers:     http.Header{"Authorization": []string{"Bearer token"}},
						Body:        []byte(`{"updated": true}`),
					},
				},
			},
			setupMocks: func(repo *testMock.MockModuleRepository) {
				updatedModule := &domainModule.Module{
					ID:          10,
					Name:        "Full Module Update",
					Description: "New description",
					OwnerId:     1,
					Webhooks: []*domainWebhook.Webhook{
						{
							ID:          1,
							Description: "Updated webhook",
							Interval:    15 * time.Second,
							Timeout:     10 * time.Second,
							URL:         url.URL{Scheme: "https", Host: "newapi.com"},
							Method:      http.MethodPut,
							Headers:     http.Header{"Authorization": []string{"Bearer token"}},
							Body:        []byte(`{"updated": true}`),
						},
					},
				}
				repo.On("UpdateModule", mock.Anything, mock.Anything, 1).Return(updatedModule, nil)
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
			result, err := service.Update(ctx, tt.module, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.module.Name, result.Name)
				assert.Equal(t, tt.module.ID, result.ID)
			}

			mockRepo.AssertExpectations(t)

			mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
			mockDLQ.AssertNotCalled(t, "Push", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
