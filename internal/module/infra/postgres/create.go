package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"

	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *PGModuleRepo) CreateModule(ctx context.Context, module domainModule.Module, userID int) (
	*domainModule.Module, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO modules (owner_id, name, description) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	err = tx.QueryRow(ctx, query, userID, module.Name, module.Description).
		Scan(&module.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, coreHttp.ErrUserNotFound
		}
		return nil, coreHttp.ErrInternal
	}
	module.OwnerId = userID

	for _, webhook := range module.Webhooks {
		webhook.ModuleID = module.ID
		if err = r.createWebhook(ctx, tx, webhook); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, coreHttp.ErrInternal
	}

	return &module, nil
}

func (r *PGModuleRepo) createWebhook(ctx context.Context, tx pgx.Tx, webhook *domainWebhook.Webhook) error {
	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return coreHttp.ErrInvalidBody
	}

	query := `INSERT INTO webhooks 
              (module_id, description, interval_s, timeout_s, url, method, headers, body) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id`

	err = tx.QueryRow(ctx, query,
		webhook.ModuleID,
		webhook.Description,
		webhook.Interval.Seconds(),
		webhook.Timeout.Seconds(),
		webhook.URL.String(),
		webhook.Method,
		headersJSON,
		webhook.Body,
	).Scan(&webhook.ID)

	if err != nil {
		return coreHttp.ErrModuleNotFound
	}

	return nil
}
