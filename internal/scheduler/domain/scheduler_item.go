package domain

import (
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"time"
)

type SchedulerItem struct {
	Attempt       int
	ScheduledTime time.Time
	Webhook       *webhookDomain.Webhook
}
