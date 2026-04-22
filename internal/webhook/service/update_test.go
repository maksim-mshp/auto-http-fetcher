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

func TestWebhookService_Update(t *testing.T) {
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
			name: "success - update webhook",
			webhook: domainWebhook.Webhook{
				ID:          10,
				Description: "Updated webhook",
				Interval:    20 * time.Second,
				Timeout:     10 * time.Second,
				URL:         url.URL{Scheme: "http", Host: "newexample.com", Path: "/webhook"},
				Method:      http.MethodPut,
				Headers:     http.Header{"Content-Type": []string{"application/json"}},
				Body:        []byte(`{"updated": true}`),
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          10,
					ModuleID:    1,
					Description: "Updated webhook",
					Interval:    20 * time.Second,
					Timeout:     10 * time.Second,
					URL:         url.URL{Scheme: "http", Host: "newexample.com", Path: "/webhook"},
					Method:      http.MethodPut,
					Headers:     http.Header{"Content-Type": []string{"application/json"}},
					Body:        []byte(`{"updated": true}`),
				}
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 1, 1).Return(expectedWebhook, nil)
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "fail - invalid user id",
			webhook: domainWebhook.Webhook{
				ID:     10,
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
				ID:     10,
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
			name: "fail - invalid webhook id",
			webhook: domainWebhook.Webhook{
				ID:     0,
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("webhook id", "webhookID is required"),
		},
		{
			name: "fail - invalid http method",
			webhook: domainWebhook.Webhook{
				ID:     10,
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
				ID:     10,
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 1, 1).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "fail - webhook not found",
			webhook: domainWebhook.Webhook{
				ID:     999,
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 1, 1).Return(nil, coreHttp.ErrWebhookNotFound)
			},
			expectedError: coreHttp.ErrWebhookNotFound,
		},
		{
			name: "fail - permission denied",
			webhook: domainWebhook.Webhook{
				ID:     10,
				Method: http.MethodGet,
				URL:    url.URL{Scheme: "http", Host: "example.com"},
			},
			moduleID: 1,
			userID:   2,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 1, 2).Return(nil, coreHttp.ErrPermissionDenied)
			},
			expectedError: coreHttp.ErrPermissionDenied,
		},
		{
			name: "success - kafka send error goes to DLQ",
			webhook: domainWebhook.Webhook{
				ID:          10,
				Description: "Update with kafka error",
				Interval:    10 * time.Second,
				Timeout:     5 * time.Second,
				URL:         url.URL{Scheme: "http", Host: "example.com"},
				Method:      http.MethodPost,
				Headers:     http.Header{},
				Body:        []byte{},
			},
			moduleID: 1,
			userID:   1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          10,
					ModuleID:    1,
					Description: "Update with kafka error",
					Interval:    10 * time.Second,
					Timeout:     5 * time.Second,
					URL:         url.URL{Scheme: "http", Host: "example.com"},
					Method:      http.MethodPost,
					Headers:     http.Header{},
					Body:        []byte{},
				}
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 1, 1).Return(expectedWebhook, nil)

				kafkaErr := errors.New("kafka timeout")
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(kafkaErr)
				dlq.On("Push", 1, mock.Anything, kafkaErr).Return()
			},
			expectedError: nil,
		},
		{
			name: "success - update webhook with all fields",
			webhook: domainWebhook.Webhook{
				ID:          20,
				Description: "Completely updated webhook",
				Interval:    60 * time.Second,
				Timeout:     30 * time.Second,
				URL:         url.URL{Scheme: "https", Host: "api.new.com", Path: "/v2/webhook"},
				Method:      http.MethodPatch,
				Headers: http.Header{
					"Authorization": []string{"Bearer newtoken"},
					"X-Version":     []string{"2.0"},
				},
				Body: []byte(`{"new": "data", "version": 2}`),
			},
			moduleID: 5,
			userID:   3,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				expectedWebhook := &domainWebhook.Webhook{
					ID:          20,
					ModuleID:    5,
					Description: "Completely updated webhook",
					Interval:    60 * time.Second,
					Timeout:     30 * time.Second,
					URL:         url.URL{Scheme: "https", Host: "api.new.com", Path: "/v2/webhook"},
					Method:      http.MethodPatch,
					Headers: http.Header{
						"Authorization": []string{"Bearer newtoken"},
						"X-Version":     []string{"2.0"},
					},
					Body: []byte(`{"new": "data", "version": 2}`),
				}
				repo.On("UpdateWebhook", mock.Anything, mock.Anything, 5, 3).Return(expectedWebhook, nil)
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
			result, err := service.Update(ctx, tt.webhook, tt.moduleID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.webhook.ID, result.ID)
				assert.Equal(t, tt.moduleID, result.ModuleID)
				assert.Equal(t, tt.webhook.Method, result.Method)
				assert.Equal(t, tt.webhook.Description, result.Description)
			}

			mockRepo.AssertExpectations(t)

			time.Sleep(50 * time.Millisecond)
			mockProducer.AssertExpectations(t)
			mockDLQ.AssertExpectations(t)
		})
	}
}
