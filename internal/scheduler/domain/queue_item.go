package domain

import (
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	"time"
)

type QueueItem struct {
	ID        int
	NextFetch time.Time
	Webhook   *domainWebhook.Webhook
	Index     int
}
