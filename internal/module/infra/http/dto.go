package http

import (
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	webhookDTO "auto-http-fetcher/internal/webhook/infra/http"
)

type ModuleDTO struct {
	ID          int                      `json:"id"`
	OwnerId     int                      `json:"owner_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Webhooks    []*webhookDTO.WebhookDTO `json:"webhooks"`
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

type ModuleListDTO struct {
	Modules []*ModuleDTO `json:"modules"`
}
