package service

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	testMock "auto-http-fetcher/internal/core/mock"

	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestModuleService_Delete(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		moduleID      int
		userID        int
		setupMocks    func(*testMock.MockModuleRepository)
		expectedError error
	}{
		{
			name:     "success - delete module",
			moduleID: 10,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("DeleteModule", mock.Anything, 10, 1).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "fail - invalid module id (zero)",
			moduleID: 0,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedError: coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:     "fail - invalid module id (negative)",
			moduleID: -5,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedError: coreHttp.NewValidationError("module id", "moduleID is required"),
		},
		{
			name:     "fail - invalid user id (zero)",
			moduleID: 10,
			userID:   0,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:     "fail - invalid user id (negative)",
			moduleID: 10,
			userID:   -3,
			setupMocks: func(repo *testMock.MockModuleRepository) {
			},
			expectedError: coreHttp.NewValidationError("user id", "userID is required"),
		},
		{
			name:     "success - module not found returns nil",
			moduleID: 999,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("DeleteModule", mock.Anything, 999, 1).Return(coreHttp.ErrModuleNotFound)
			},
			expectedError: nil,
		},
		{
			name:     "fail - repository returns other error",
			moduleID: 10,
			userID:   1,
			setupMocks: func(repo *testMock.MockModuleRepository) {
				repo.On("DeleteModule", mock.Anything, 10, 1).Return(errors.New("database connection failed"))
			},
			expectedError: errors.New("database connection failed"),
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
			err := service.Delete(ctx, tt.moduleID, tt.userID)

			if tt.expectedError != nil {
				t.Log("name", tt.name)
				t.Log("error", err)
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)

			mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
			mockDLQ.AssertNotCalled(t, "Push", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
