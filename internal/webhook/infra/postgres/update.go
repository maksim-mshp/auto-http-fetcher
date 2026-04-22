package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/webhook/domain"
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
)

func (r *PGWebhookRepo) UpdateWebhook(ctx context.Context, webhook domain.Webhook, moduleID, userID int) (*domain.Webhook, error) {
	checkQuery := `SELECT EXISTS(SELECT 1 FROM modules WHERE id = $1 AND owner_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, checkQuery, moduleID, userID).Scan(&exists)
	if err != nil {
		log.Println(1)
		return nil, coreHttp.ErrInternal
	}
	if !exists {
		return nil, coreHttp.ErrModuleNotFound
	}

	query := `UPDATE webhooks 
              SET description = $1, interval_s = $2, timeout_s = $3, 
                  url = $4, method = $5, headers = $6, body = $7, updated_at = NOW()
              WHERE id = $8 AND module_id = $9
              RETURNING id`

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return nil, coreHttp.ErrInvalidBody
	}

	err = r.pool.QueryRow(ctx, query,
		webhook.Description,
		webhook.Interval.Seconds(),
		webhook.Timeout.Seconds(),
		webhook.URL.String(),
		webhook.Method,
		headersJSON,
		webhook.Body,
		webhook.ID,
		moduleID,
	).Scan(&webhook.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, coreHttp.ErrWebhookNotFound
		}
		log.Println(1)
		return nil, coreHttp.ErrInternal
	}
	webhook.ModuleID = moduleID
	return &webhook, nil
}
