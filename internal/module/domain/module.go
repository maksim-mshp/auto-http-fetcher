package domain

import "auto-http-fetcher/internal/webhook/domain"

type Module struct {
	ID          int
	OwnerId     int
	Name        string
	Description string
	Webhooks    []*domain.Webhook
}
