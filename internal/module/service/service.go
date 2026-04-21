package service

import (
	coreKafka "auto-http-fetcher/internal/core/kafka"
	"auto-http-fetcher/internal/module/infra/kafka/dlq"

	"log/slog"
)

type ModuleService struct {
	logger     *slog.Logger
	moduleRepo ModuleRepository
	kafka      *coreKafka.Producer
	dlq        *dlq.DeadLetterQueue
}

func NewModuleService(logger *slog.Logger, kafka *coreKafka.Producer, dlq *dlq.DeadLetterQueue,
	repo ModuleRepository) *ModuleService {

	return &ModuleService{
		logger:     logger,
		moduleRepo: repo,
		kafka:      kafka,
		dlq:        dlq,
	}
}
