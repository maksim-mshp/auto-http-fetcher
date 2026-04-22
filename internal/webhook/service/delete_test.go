package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	testMock "auto-http-fetcher/internal/core/mock"

	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhookService_Delete(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		webhookID     int
		moduleID      int
		userID        int
		setupMocks    func(*testMock.MockWebhookRepository, *testMock.MockProducer, *testMock.MockDLQ)
		expectedError error
	}{
		{
			name:      "success - delete webhook",
			webhookID: 10,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("DeleteWebhook", mock.Anything, 10, 1, 1).Return(nil)
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "fail - invalid user id (zero)",
			webhookID: 10,
			moduleID:  1,
			userID:    0,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:      "fail - invalid user id (negative)",
			webhookID: 10,
			moduleID:  1,
			userID:    -5,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:      "fail - invalid module id (zero)",
			webhookID: 10,
			moduleID:  0,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:      "fail - invalid module id (negative)",
			webhookID: 10,
			moduleID:  -3,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:      "fail - invalid webhook id (zero)",
			webhookID: 0,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("webhook id", "webhookID is required"),
		},
		{
			name:      "fail - invalid webhook id (negative)",
			webhookID: -10,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
			},
			expectedError: coreHttp.NewValidationError("webhook id", "webhookID is required"),
		},
		{
			name:      "success - webhook not found returns nil",
			webhookID: 999,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("DeleteWebhook", mock.Anything, 999, 1, 1).Return(coreHttp.ErrWebhookNotFound)
			},
			expectedError: nil,
		},
		{
			name:      "fail - repository error",
			webhookID: 10,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("DeleteWebhook", mock.Anything, 10, 1, 1).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:      "success - kafka send error goes to DLQ",
			webhookID: 10,
			moduleID:  1,
			userID:    1,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("DeleteWebhook", mock.Anything, 10, 1, 1).Return(nil)

				kafkaErr := errors.New("kafka unavailable")
				producer.On("SendMessage", mock.Anything, 1, mock.Anything).Return(kafkaErr)
				dlq.On("Push", 1, mock.Anything, kafkaErr).Return()
			},
			expectedError: nil,
		},
		{
			name:      "success - delete with multiple webhooks",
			webhookID: 20,
			moduleID:  5,
			userID:    3,
			setupMocks: func(repo *testMock.MockWebhookRepository, producer *testMock.MockProducer, dlq *testMock.MockDLQ) {
				repo.On("DeleteWebhook", mock.Anything, 20, 5, 3).Return(nil)
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
			err := service.Delete(ctx, tt.webhookID, tt.moduleID, tt.userID)

			if tt.expectedError != nil {
				t.Log("name", tt.name)
				t.Log("error", err)
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)

			time.Sleep(50 * time.Millisecond)
			mockProducer.AssertExpectations(t)
			mockDLQ.AssertExpectations(t)
		})
	}
}
