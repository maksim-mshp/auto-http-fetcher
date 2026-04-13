package domain

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, r *Response) error
	FindByID(ctx context.Context, id string) (*Response, error)
	FindByWebhookID(ctx context.Context, webhookID string) ([]*Response, error)
}
