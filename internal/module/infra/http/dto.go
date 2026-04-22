package http

import (
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	webhookDTO "auto-http-fetcher/internal/webhook/infra/http"
)

type ModuleRequestResponse struct {
	Module ModuleDTO `json:"module"`
}

type ModuleList struct {
	Modules []*ModuleDTO `json:"modules"`
}

type ModuleDTO struct {
	ID          int                      `json:"id"`
	OwnerId     int                      `json:"owner_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Webhooks    []*webhookDTO.WebhookDTO `json:"webhooks"`
}

func (m *ModuleDTO) ToDomain() (*domainModule.Module, error) {
	module := domainModule.Module{
		ID:          m.ID,
		OwnerId:     m.OwnerId,
		Name:        m.Name,
		Description: m.Description,
		Webhooks:    make([]*domainWebhook.Webhook, len(m.Webhooks)),
	}
	for i, webhook := range m.Webhooks {
		webhookToDomain, err := webhook.ToDomain()
		if err != nil {
			return nil, err
		}
		module.Webhooks[i] = webhookToDomain
	}
	return &module, nil
}

func ModuleToDTO(m *domainModule.Module) *ModuleDTO {
	var module = ModuleDTO{
		ID:          m.ID,
		OwnerId:     m.OwnerId,
		Name:        m.Name,
		Description: m.Description,
		Webhooks:    make([]*webhookDTO.WebhookDTO, len(m.Webhooks)),
	}
	for i, webhook := range m.Webhooks {
		module.Webhooks[i] = webhookDTO.WebhookToDTO(webhook)
	}
	return &module
}
