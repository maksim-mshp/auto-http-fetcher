package domain

import (
	"net/http"
	"net/url"
	"time"
)

type WebhookType string

const (
	ManualWebhook    WebhookType = "Manual"
	ScheduledWebhook WebhookType = "Scheduled"
)

type Webhook struct {
	ID          int
	Description string

	Type WebhookType

	Interval time.Duration
	Timeout  time.Duration

	URL     url.URL
	Method  string
	Headers http.Header
	Body    []byte
}
