package service

import (
	"auto-http-fetcher/internal/module/infra/kafka"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"
	"context"
	"time"
)

func (s *WebhookService) toKafka(action string, userID int, webhook *domainWebhook.Webhook) {
	go func() {
		ctxNew, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		message := kafka.WebhookKafkaDTO{
			Action:      action,
			ID:          webhook.ID,
			Description: webhook.Description,
			Interval:    int(webhook.Interval.Seconds()),
			Timeout:     int(webhook.Timeout.Seconds()),
			URL:         webhook.URL.String(),
			Method:      webhook.Method,
			Headers:     webhook.Headers,
			Body:        webhook.Body,
		}

		err := s.kafka.SendMessage(ctxNew, userID, message)
		if err != nil {
			s.dlq.Push(userID, message, err)
		}
	}()
}
