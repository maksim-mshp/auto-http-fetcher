package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/webhook/domain"

	"context"
	"encoding/json"
)

func (r *PGWebhookRepo) CreateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error) {
	checkQuery := `SELECT EXISTS(SELECT 1 FROM modules WHERE id = $1 AND owner_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, checkQuery, moduleID, userID).Scan(&exists)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	if !exists {
		return nil, coreHttp.ErrModuleNotFound
	}

	query := `INSERT INTO webhooks 
              (module_id, description, interval_s, timeout_s, url, method, headers, body) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id`

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return nil, coreHttp.ErrInvalidBody
	}

	webhook.ModuleID = moduleID
	err = r.pool.QueryRow(ctx, query,
		moduleID,
		webhook.Description,
		webhook.Interval.Seconds(),
		webhook.Timeout.Seconds(),
		webhook.URL.String(),
		webhook.Method,
		headersJSON,
		webhook.Body,
	).Scan(&webhook.ID)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	webhook.ModuleID = moduleID
	return &webhook, nil
}
