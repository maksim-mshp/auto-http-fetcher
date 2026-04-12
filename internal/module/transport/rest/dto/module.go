package dto

import (
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	"net/http"
	"net/url"
	"time"
)

type ModuleDTO struct {
	ID          int           `json:"module_id"`
	OwnerId     int           `json:"module_owner_id"`
	Name        string        `json:"module_name"`
	Description string        `json:"module_description"`
	Webhooks    []*WebhookDTO `json:"module_webhooks"`
}

func (m *ModuleDTO) ToDomain() domainModule.Module {
	module := domainModule.Module{
		ID:          m.ID,
		OwnerId:     m.OwnerId,
		Name:        m.Name,
		Description: m.Description,
		Webhooks:    make([]*domainWebhook.Webhook, len(m.Webhooks)),
	}
	for i, webhook := range m.Webhooks {
		module.Webhooks[i] = webhook.ToDomain()
	}
	return module
}

type WebhookDTO struct {
	ID          int    `json:"webhook_id"`
	Description string `json:"webhook_description"`

	Interval time.Duration `json:"webhook_interval"`
	Timeout  time.Duration `json:"webhook_timeout"`

	URL     url.URL     `json:"webhook_url"`
	Method  string      `json:"webhook_method"`
	Headers http.Header `json:"webhook_headers"`
	Body    []byte      `json:"webhook_body"`
}

func (w *WebhookDTO) ToDomain() *domainWebhook.Webhook {
	return &domainWebhook.Webhook{
		ID:          w.ID,
		Description: w.Description,
		Interval:    w.Interval,
		Timeout:     w.Timeout,
		Method:      w.Method,
		Headers:     w.Headers,
		Body:        w.Body,
	}
}

type ModuleListDTO struct {
	Modules []*ModuleDTO `json:"modules"`
}
