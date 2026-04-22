package service

import (
	coreKafka "auto-http-fetcher/internal/core/kafka"
	"auto-http-fetcher/internal/module/infra/kafka/dlq"
	"log/slog"
)

type WebhookService struct {
	logger      *slog.Logger
	webhookRepo WebhookRepository
	kafka       coreKafka.ProducerInterface
	dlq         dlq.DLQ
}

func NewWebhookService(logger *slog.Logger, kafka *coreKafka.Producer, dlq *dlq.DeadLetterQueue,
	repo WebhookRepository) *WebhookService {

	return &WebhookService{
		logger:      logger,
		webhookRepo: repo,
		kafka:       kafka,
		dlq:         dlq,
	}
}
