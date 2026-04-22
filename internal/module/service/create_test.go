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

func TestModuleService_Create(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		userID        int
		module        domainModule.Module
		setupMocks    func(*testMock.MockModuleRepository, *testMock.MockProducer, *testMock.MockDLQ)
		expectedError error
	}{
		{
			name:   "success - valid module",
			userID: 1,
			module: domainModule.Module{
				Name: "Test Module",
				Webhooks: []*domainWebhook.Webhook{
					{
						ID:          1,
						Description: "Test webhook",
						Interval:    10 * time.Second,
						Timeout:     5 * time.Second,
						URL:         url.URL{Scheme: "http", Host: "example.com", Path: "/webhook"},
						Method:      http.MethodPost,
						Headers:     http.Header{"Content-Type": []string{"application/json"}},
						Body:        []byte(`{"test": true}`),
					},
				},
			},
			setupMocks: func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedModule := &domainModule.Module{
					Name: "Test Module",
					Webhooks: []*domainWebhook.Webhook{
						{
							ID:          1,
							Description: "Test webhook",
							Interval:    10 * time.Second,
							Timeout:     5 * time.Second,
							URL:         url.URL{Scheme: "http", Host: "example.com", Path: "/webhook"},
							Method:      http.MethodPost,
							Headers:     http.Header{"Content-Type": []string{"application/json"}},
							Body:        []byte(`{"test": true}`),
						},
					},
				}
				repo.On("CreateModule", mock.Anything, mock.Anything, 1).Return(expectedModule, nil)

				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "fail - invalid user id",
			userID: 0,
			module: domainModule.Module{
				Name: "Test Module",
			},
			setupMocks:    func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:   "fail - empty module name",
			userID: 1,
			module: domainModule.Module{
				Name: "",
			},
			setupMocks:    func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {},
			expectedError: coreHttp.NewValidationError("name", "name is required"),
		},
		{
			name:   "fail - invalid http method",
			userID: 1,
			module: domainModule.Module{
				Name: "Test Module",
				Webhooks: []*domainWebhook.Webhook{
					{
						Method: "INVALID_METHOD",
						URL:    url.URL{Scheme: "http", Host: "example.com"},
					},
				},
			},
			setupMocks:    func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {},
			expectedError: coreHttp.ErrInvalidBody,
		},
		{
			name:   "fail - repository error",
			userID: 1,
			module: domainModule.Module{
				Name: "Test Module",
			},
			setupMocks: func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("CreateModule", mock.Anything, mock.Anything, 1).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:   "success - kafka send error goes to DLQ",
			userID: 1,
			module: domainModule.Module{
				Name: "Test Module",
				Webhooks: []*domainWebhook.Webhook{
					{
						ID:       1,
						Method:   http.MethodGet,
						URL:      url.URL{Scheme: "http", Host: "example.com"},
						Headers:  http.Header{},
						Body:     []byte{},
						Interval: 10 * time.Second,
						Timeout:  5 * time.Second,
					},
				},
			},
			setupMocks: func(repo *testMock.MockModuleRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedModule := &domainModule.Module{
					Name: "Test Module",
					Webhooks: []*domainWebhook.Webhook{
						{
							ID:       1,
							Method:   http.MethodGet,
							URL:      url.URL{Scheme: "http", Host: "example.com"},
							Headers:  http.Header{},
							Body:     []byte{},
							Interval: 10 * time.Second,
							Timeout:  5 * time.Second,
						},
					},
				}
				repo.On("CreateModule", mock.Anything, mock.Anything, 1).Return(expectedModule, nil)

				kafkaErr := errors.New("kafka unavailable")
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(kafkaErr)

				dlq.On("Push", 1, mock.Anything, kafkaErr).Return()
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testMock.MockModuleRepository)
			mockProducer := new(testMock.MockProducer)
			mockDLQ := new(testMock.MockDLQ)

			tt.setupMocks(mockRepo, mockProducer, mockDLQ)

			service := &ModuleService{
				logger:     logger,
				moduleRepo: mockRepo,
				kafka:      mockProducer,
				dlq:        mockDLQ,
			}

			ctx := context.Background()
			result, err := service.Create(ctx, tt.module, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)

			time.Sleep(100 * time.Millisecond)
			mockProducer.AssertExpectations(t)
			mockDLQ.AssertExpectations(t)
		})
	}
}
