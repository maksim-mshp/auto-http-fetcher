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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestModuleService_List(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name            string
		userID          int
		setupMocks      func(*testMock.MockModuleRepository)
		expectedModules []*domainModule.Module
		expectedError   error
	}{
		{
			name:   "success - get modules list",
			userID: 1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				expectedModules := []*domainModule.Module{
					{
						ID:          1,
						Name:        "Module 1",
						Description: "First module",
						OwnerId:     1,
						Webhooks: []*domainWebhook.Webhook{
							{
								ID:      1,
								Method:  http.MethodGet,
								URL:     url.URL{Scheme: "http", Host: "api1.com"},
								Headers: http.Header{},
								Body:    []byte{},
							},
						},
					},
					{
						ID:          2,
						Name:        "Module 2",
						Description: "Second module",
						OwnerId:     1,
						Webhooks:    []*domainWebhook.Webhook{},
					},
					{
						ID:          3,
						Name:        "Module 3",
						Description: "Third module",
						OwnerId:     1,
						Webhooks: []*domainWebhook.Webhook{
							{
								ID:      2,
								Method:  http.MethodPost,
								URL:     url.URL{Scheme: "https", Host: "api2.com"},
								Headers: http.Header{"X-Token": []string{"123"}},
								Body:    []byte(`{"data": "test"}`),
							},
						},
					},
				}
				repo.On("GetModuleList", mock.Anything, 1).Return(expectedModules, nil)
			},
			expectedModules: []*domainModule.Module{
				{ID: 1, Name: "Module 1", OwnerId: 1},
				{ID: 2, Name: "Module 2", OwnerId: 1},
				{ID: 3, Name: "Module 3", OwnerId: 1},
			},
			expectedError: nil,
		},
		{
			name:   "success - empty modules list",
			userID: 1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModuleList", mock.Anything, 1).Return([]*domainModule.Module{}, nil)
			},
			expectedModules: []*domainModule.Module{},
			expectedError:   nil,
		},
		{
			name:   "success - nil modules list",
			userID: 1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModuleList", mock.Anything, 1).Return(nil, nil)
			},
			expectedModules: nil,
			expectedError:   nil,
		},
		{
			name:   "fail - invalid user id (zero)",
			userID: 0,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModules: nil,
			expectedError:   coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:   "fail - invalid user id (negative)",
			userID: -5,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedModules: nil,
			expectedError:   coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:   "fail - repository error",
			userID: 1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModuleList", mock.Anything, 1).Return(nil, errors.New("database error"))
			},
			expectedModules: nil,
			expectedError:   errors.New("database error"),
		},
		{
			name:   "fail - permission denied",
			userID: 2,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("GetModuleList", mock.Anything, 2).Return(nil, coreHttp.ErrPermissionDenied)
			},
			expectedModules: nil,
			expectedError:   coreHttp.ErrPermissionDenied,
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
			result, err := service.List(ctx, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedModules == nil {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, len(tt.expectedModules), len(result))
					for i, expected := range tt.expectedModules {
						assert.Equal(t, expected.ID, result[i].ID)
						assert.Equal(t, expected.Name, result[i].Name)
						assert.Equal(t, expected.OwnerId, result[i].OwnerId)
					}
				}
			}

			mockRepo.AssertExpectations(t)

			mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
			mockDLQ.AssertNotCalled(t, "Push", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
