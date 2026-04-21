package loader

import (
	kafkaProducer "auto-http-fetcher/internal/core/kafka"
	"auto-http-fetcher/internal/module/infra/kafka"
	"auto-http-fetcher/internal/module/infra/kafka/dlq"
	"auto-http-fetcher/internal/webhook/domain"

	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookLoader - воркер который при старте module сервиса берет все вебхуки, что есть и в кафку кидает (для scheduler)
type WebhookLoader struct {
	pool            *pgxpool.Pool
	kafka           *kafkaProducer.Producer
	logger          *slog.Logger
	deadLetterQueue *dlq.DeadLetterQueue
}

func NewWebhookLoader(logger *slog.Logger, pool *pgxpool.Pool, kafka *kafkaProducer.Producer,
	d *dlq.DeadLetterQueue) *WebhookLoader {

	return &WebhookLoader{
		pool:            pool,
		kafka:           kafka,
		logger:          logger,
		deadLetterQueue: d,
	}
}

func (loader *WebhookLoader) Load(ctx context.Context) error {
	loader.logger.Info("starting webhook loader")

	webhooks, err := loader.loadFromDB(ctx)
	if err != nil {
		return err
	}

	if len(webhooks) == 0 {
		loader.logger.Info("no webhooks found in database")
		return nil
	}

	loader.logger.Info("loaded webhooks from database", "count", len(webhooks))
	const userID = 0
	for _, webhook := range webhooks {
		const action = "create"
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

		if err := loader.kafka.SendMessage(ctx, userID, message); err != nil {
			loader.deadLetterQueue.Push(userID, message, err)

			loader.logger.Warn("webhook sent to dead letter queue instead kafka", "id", userID,
				"webhook_id", webhook.ID, "err", err)

		} else {
			loader.logger.Debug("webhook sent to Kafka",
				"webhook_id", webhook.ID,
				"module_id", webhook.ModuleID)
		}
	}

	loader.logger.Info("webhook loading completed")
	return nil
}

func (loader *WebhookLoader) loadFromDB(ctx context.Context) ([]*domain.Webhook, error) {
	query := `
		SELECT 
			w.id, w.module_id, w.description, 
			w.interval_s, w.timeout_s,
			w.url, w.method, w.headers, w.body
		FROM webhooks w
		ORDER BY w.id ASC
	`

	rows, err := loader.pool.Query(ctx, query)
	if err != nil {
		loader.logger.Error("failed to query webhooks", "error", err)
		return nil, err
	}
	defer rows.Close()

	var webhooks []*domain.Webhook

	for rows.Next() {
		webhook := &domain.Webhook{}
		var headersJSON []byte
		var urlStr string
		var intervalS, timeoutS int64

		err = rows.Scan(
			&webhook.ID,
			&webhook.ModuleID,
			&webhook.Description,
			&intervalS,
			&timeoutS,
			&urlStr,
			&webhook.Method,
			&headersJSON,
			&webhook.Body,
		)
		if err != nil {
			loader.logger.Error("failed to scan webhook", "error", err)
			continue
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			loader.logger.Error("failed to parse URL", "webhook_id", webhook.ID, "url", urlStr, "error", err)
			continue
		}
		webhook.URL = *parsedURL

		webhook.Interval = time.Duration(intervalS) * time.Second
		webhook.Timeout = time.Duration(timeoutS) * time.Second

		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
				loader.logger.Warn("failed to unmarshal headers", "webhook_id", webhook.ID, "error", err)
				webhook.Headers = make(http.Header)
			}
		} else {
			webhook.Headers = make(http.Header)
		}

		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		loader.logger.Error("rows iteration error", "error", err)
		return nil, err
	}

	return webhooks, nil
}
