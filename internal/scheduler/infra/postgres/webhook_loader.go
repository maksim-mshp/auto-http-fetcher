package postgres

import (
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookLoader struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewWebhookLoader(pool *pgxpool.Pool, logger *slog.Logger) *WebhookLoader {
	return &WebhookLoader{
		pool:   pool,
		logger: logger,
	}
}

func (l *WebhookLoader) Load(ctx context.Context) ([]*webhookDomain.Webhook, error) {
	query := `
		SELECT id, module_id, description, interval_s, timeout_s, url, method, headers, body
		FROM webhooks
		ORDER BY id ASC
	`

	rows, err := l.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*webhookDomain.Webhook
	for rows.Next() {
		webhook := &webhookDomain.Webhook{}
		var headersJSON []byte
		var urlStr string
		var intervalS, timeoutS int64

		if err = rows.Scan(
			&webhook.ID,
			&webhook.ModuleID,
			&webhook.Description,
			&intervalS,
			&timeoutS,
			&urlStr,
			&webhook.Method,
			&headersJSON,
			&webhook.Body,
		); err != nil {
			l.logger.Error("scan scheduled webhook failed", "err", err)
			continue
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			l.logger.Error("parse scheduled webhook url failed", "id", webhook.ID, "url", urlStr, "err", err)
			continue
		}
		webhook.URL = *parsedURL
		webhook.Interval = time.Duration(intervalS) * time.Second
		webhook.Timeout = time.Duration(timeoutS) * time.Second

		if len(headersJSON) == 0 {
			webhook.Headers = http.Header{}
		} else if err = json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
			l.logger.Error("unmarshal scheduled webhook headers failed", "id", webhook.ID, "err", err)
			webhook.Headers = http.Header{}
		}

		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return webhooks, nil
}
