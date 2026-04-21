package kafka

import (
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

type Scheduler interface {
	AddWebhook(wh *webhookDomain.Webhook)
	UpsertWebhook(wh *webhookDomain.Webhook)
	DeleteWebhook(id int)
}

type Consumer struct {
	group     sarama.ConsumerGroup
	topics    []string
	scheduler Scheduler
	logger    *slog.Logger
}

func NewConsumer(
	brokers []string,
	groupID string,
	topics []string,
	scheduler Scheduler,
	logger *slog.Logger,
) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		group:     group,
		topics:    topics,
		scheduler: scheduler,
		logger:    logger,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	handler := &consumerGroupHandler{
		scheduler: c.scheduler,
		logger:    c.logger,
	}

	for ctx.Err() == nil {
		if err := c.group.Consume(ctx, c.topics, handler); err != nil {
			return err
		}
	}

	return ctx.Err()
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

type consumerGroupHandler struct {
	scheduler Scheduler
	logger    *slog.Logger
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.handleMessage(message); err != nil {
			h.logger.Error("scheduler kafka message failed",
				"topic", message.Topic,
				"partition", message.Partition,
				"offset", message.Offset,
				"err", err,
			)
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func (h *consumerGroupHandler) handleMessage(message *sarama.ConsumerMessage) error {
	var payload WebhookPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return fmt.Errorf("decode webhook payload: %w", err)
	}

	action := strings.ToLower(strings.TrimSpace(payload.Action))
	switch action {
	case ActionCreate:
		wh, err := payload.ToDomain()
		if err != nil {
			return err
		}
		h.scheduler.AddWebhook(wh)
	case ActionUpdate:
		wh, err := payload.ToDomain()
		if err != nil {
			return err
		}
		h.scheduler.UpsertWebhook(wh)
	case ActionDelete:
		if payload.ID <= 0 {
			return fmt.Errorf("delete webhook id is required")
		}
		h.scheduler.DeleteWebhook(payload.ID)
	default:
		return fmt.Errorf("unknown webhook action %q", payload.Action)
	}

	h.logger.Info("scheduler kafka message handled", "action", action, "webhook_id", payload.ID)
	return nil
}

func (p WebhookPayload) ToDomain() (*webhookDomain.Webhook, error) {
	if p.ID <= 0 {
		return nil, fmt.Errorf("webhook id is required")
	}

	parsedURL, err := url.Parse(p.URL)
	if err != nil {
		return nil, fmt.Errorf("parse webhook url: %w", err)
	}

	headers := http.Header{}
	for key, values := range p.Headers {
		headers[key] = append([]string(nil), values...)
	}

	return &webhookDomain.Webhook{
		ID:          p.ID,
		Description: p.Description,
		Interval:    time.Duration(p.Interval) * time.Second,
		Timeout:     time.Duration(p.Timeout) * time.Second,
		URL:         *parsedURL,
		Method:      p.Method,
		Headers:     headers,
		Body:        append([]byte(nil), p.Body...),
	}, nil
}
