package http

import (
	"auto-http-fetcher/internal/webhook/domain"

	"net/url"
	"time"
)

type WebhookDTO struct {
	ID          int    `json:"id" example:"1"`
	ModuleID    int    `json:"module_id" example:"1"`
	Description string `json:"description" example:"Ping API"`

	Interval string `json:"interval" example:"1m0s"`
	Timeout  string `json:"timeout" example:"5s"`

	URL     string              `json:"url" example:"https://example.com/health"`
	Method  string              `json:"method" enums:"GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE" example:"GET"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body" swaggertype:"string" format:"byte" example:""`
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
		ModuleID:    w.ModuleID,
		Description: w.Description,
		Interval:    interval,
		Timeout:     timeout,
		URL:         *parsedURL,
		Method:      w.Method,
		Headers:     w.Headers,
		Body:        w.Body,
	}, nil
}

func WebhookToDTO(w *domain.Webhook) *WebhookDTO {
	return &WebhookDTO{
		ID:          w.ID,
		ModuleID:    w.ModuleID,
		Description: w.Description,
		Interval:    w.Interval.String(),
		Timeout:     w.Timeout.String(),
		Method:      w.Method,
		Headers:     w.Headers,
		Body:        w.Body,
		URL:         w.URL.String(),
	}
}

type WebhookDTORequestResponse struct {
	Webhook *WebhookDTO `json:"webhook"`
}
