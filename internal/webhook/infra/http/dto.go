package http

import (
	"auto-http-fetcher/internal/webhook/domain"
	"net/http"
	"net/url"
	"time"
)

type WebhookDTO struct {
	ID          int    `json:"id"`
	Description string `json:"description"`

	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`

	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

func (w *WebhookDTO) ToDomain() (*domain.Webhook, error) {
	interval, err := time.ParseDuration(w.Interval)
	if err != nil {
		return nil, err
	}
	timeout, err := time.ParseDuration(w.Timeout)
	if err != nil {
		return nil, err
	}
	parsedURL, err := url.Parse(w.URL)
	if err != nil {
		return nil, err
	}

	return &domain.Webhook{
		ID:          w.ID,
		Description: w.Description,
		Interval:    interval * time.Second,
		Timeout:     timeout * time.Second,
		URL:         *parsedURL,
		Method:      w.Method,
		Headers:     w.Headers,
		Body:        w.Body,
	}, nil
}
