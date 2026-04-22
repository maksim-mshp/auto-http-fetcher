package mock

import (
	kafka2 "auto-http-fetcher/internal/module/infra/kafka"

	"github.com/stretchr/testify/mock"
)

type MockDLQ struct {
	mock.Mock
}

func (m *MockDLQ) Push(userID int, msg kafka2.WebhookKafkaDTO, err error) {
	m.Called(userID, msg, err)
}
