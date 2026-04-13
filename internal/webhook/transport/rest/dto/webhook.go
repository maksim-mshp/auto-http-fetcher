package dto

import (
	"auto-http-fetcher/internal/webhook/domain"
	"net/http"
	"net/url"
	"time"
)

type WebhookDTO struct {
	ID          int    `json:"id"`
	Description string `json:"description"`

	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`

	URL     url.URL     `json:"url"`
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

func (w *WebhookDTO) ToDomain() *domain.Webhook {
	return &domain.Webhook{
		ID:          w.ID,
		Description: w.Description,
		Interval:    w.Interval,
		Timeout:     w.Timeout,
		Method:      w.Method,
		Headers:     w.Headers,
		Body:        w.Body,
	}
}
