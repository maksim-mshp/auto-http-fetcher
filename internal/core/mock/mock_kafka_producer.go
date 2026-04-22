package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(ctx context.Context, userID int, message any) error {
	args := m.Called(ctx, userID, message)
	return args.Error(0)
}
