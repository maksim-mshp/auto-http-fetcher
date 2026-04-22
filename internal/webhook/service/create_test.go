package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	testMock "auto-http-fetcher/internal/core/mock"
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

func TestWebhookService_Create(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		webhook       domainWebhook.Webhook
		moduleID      int
		userID        int
		setupMocks    func(*testMock.MockWebhookRepository, *testMock.MockProducer, *testMock.MockDLQ)
		expectedError error
	}{
		{
			name: "success - create webhook",
			webhook: domainWebhook.Webhook{
				Description: "Test webhook",
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				URL:         url.URL{Scheme: "http", Host: "example.com", Path: "/webhook"},
				Method:      http.MethodPost,
				Headers:     http.Header{"Content-Type": []string{"application/json"}},
				Body:        []byte(`{"test": true}`),
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          10,
					ModuleID:    1,
					Description: "Test webhook",
					Interval:    10 * time.Second,
					Timeout:     5 * time.Second,
					URL:         url.URL{Scheme: "http", Host: "example.com", Path: "/webhook"},
					Method:      http.MethodPost,
					Headers:     http.Header{"Content-Type": []string{"application/json"}},
					Body:        []byte(`{"test": true}`),
				}
				repo.On("CreateWebhook", mock.Anything, mock.Anything, 1, 1).Return(expectedWebhook, nil)
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "fail - invalid user id",
			webhook: domainWebhook.Webhook{
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   0,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name: "fail - invalid module id",
			webhook: domainWebhook.Webhook{
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 0,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name: "fail - invalid http method",
			webhook: domainWebhook.Webhook{
				Method: "INVALID_METHOD",
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.ErrInvalidBody,
		},
		{
			name: "fail - repository error",
			webhook: domainWebhook.Webhook{
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("CreateWebhook", mock.Anything, mock.Anything, 1, 1).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "success - kafka send error goes to DLQ",
			webhook: domainWebhook.Webhook{
				Description: "Test webhook",
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				URL:         url.URL{Scheme: "http", Host: "example.com"},
				Method:      http.MethodGet,
				Headers:     http.Header{},
				Body:        []byte{},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          10,
					ModuleID:    1,
					Description: "Test webhook",
					Interval:    10 * time.Second,
					Timeout:     5 * time.Second,
					URL:         url.URL{Scheme: "http", Host: "example.com"},
					Method:      http.MethodGet,
					Headers:     http.Header{},
					Body:        []byte{},
				}
				repo.On("CreateWebhook", mock.Anything, mock.Anything, 1, 1).Return(expectedWebhook, nil)

				kafkaErr := errors.New("kafka unavailable")
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(kafkaErr)
				dlq.On("Push", 1, mock.Anything, kafkaErr).Return()
			},
			expectedError: nil,
		},
		{
			name: "success - webhook with all fields",
			webhook: domainWebhook.Webhook{
				Description: "Full webhook",
				Interval:    30 * time.Second,
				Timeout:     15 * time.Second,
				URL:         url.URL{Scheme: "https", Host: "api.example.com", Path: "/v1/webhook"},
				Method:      http.MethodPut,
				Headers: http.Header{
					"Authorization": []string{"Bearer token123"},
					"X-Custom":      []string{"custom-value"},
				},
				Body: []byte(`{"key": "value", "nested": {"field": true}}`),
			},
			moduleID: 5,
			userID:   3,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          20,
					ModuleID:    5,
					Description: "Full webhook",
					Interval:    30 * time.Second,
					Timeout:     15 * time.Second,
					URL:         url.URL{Scheme: "https", Host: "api.example.com", Path: "/v1/webhook"},
					Method:      http.MethodPut,
					Headers: http.Header{
						"Authorization": []string{"Bearer token123"},
						"X-Custom":      []string{"custom-value"},
					},
					Body: []byte(`{"key": "value", "nested": {"field": true}}`),
				}
				repo.On("CreateWebhook", mock.Anything, mock.Anything, 5, 3).Return(expectedWebhook, nil)
				producer.On("SendMessage", mock.Anything, 3, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(testMock.MockWebhookRepository)
			mockProducer := new(testMock.MockProducer)
			mockDLQ := new(testMock.MockDLQ)

			tt.setupMocks(mockRepo, mockProducer, mockDLQ)

			service := &WebhookService{
				logger:      logger,
				webhookRepo: mockRepo,
				kafka:       mockProducer,
				dlq:         mockDLQ,
			}

			ctx := context.Background()
			result, err := service.Create(ctx, tt.webhook, tt.moduleID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.moduleID, result.ModuleID)
				assert.Equal(t, tt.webhook.Method, result.Method)
			}

			mockRepo.AssertExpectations(t)

			time.Sleep(50 * time.Millisecond)
			mockProducer.AssertExpectations(t)
			mockDLQ.AssertExpectations(t)
		})
	}
}
