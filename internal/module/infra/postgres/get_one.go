package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"

	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *PGModuleRepo) GetModule(ctx context.Context, moduleID, userID int) (*domainModule.Module, error) {

	query := `SELECT id, owner_id, name, description
              FROM modules WHERE id = $1 AND owner_id = $2`

	var module domainModule.Module
	err := r.pool.QueryRow(ctx, query, moduleID, userID).Scan(
		&module.ID, &module.OwnerId, &module.Name, &module.Description)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, coreHttp.ErrModuleNotFound
		}
		log.Println("1", err)
		return nil, coreHttp.ErrInternal
	}

	webhooks, err := r.getWebhooksByModuleID(ctx, module.ID)
	if err != nil {
		log.Println("2", err)
		return nil, coreHttp.ErrInternal
	}
	module.Webhooks = webhooks

	return &module, nil
}

func (r *PGModuleRepo) getWebhooksByModuleID(ctx context.Context, moduleID int) ([]*domainWebhook.Webhook, error) {
	return r.getWebhooksByModuleIDs(ctx, []int{moduleID})
}

func (r *PGModuleRepo) getWebhooksByModuleIDs(ctx context.Context, moduleIDs []int) ([]*domainWebhook.Webhook, error) {
	query := `SELECT id, module_id, description, interval_s, timeout_s, url, method, headers, body
              FROM webhooks WHERE module_id = ANY($1)`

	rows, err := r.pool.Query(ctx, query, moduleIDs)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	defer rows.Close()

	var webhooks []*domainWebhook.Webhook
	for rows.Next() {
		webhook := &domainWebhook.Webhook{}
		var headersJSON []byte
		var urlStr string
		var interval, timeout int64

		err = rows.Scan(
			&webhook.ID,
			&webhook.ModuleID,
			&webhook.Description,
			&interval,
			&timeout,
			&urlStr,
			&webhook.Method,
			&headersJSON,
			&webhook.Body,
		)
		if err != nil {
			log.Println("3", err)
			return nil, coreHttp.ErrInternal
		}

		webhook.Interval = time.Duration(interval) * time.Second
		webhook.Timeout = time.Duration(timeout) * time.Second

		urlParse, err := url.Parse(urlStr)
		if err != nil {
			return nil, coreHttp.ErrInternal
		}
		webhook.URL = *urlParse

		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
				return nil, coreHttp.ErrInvalidBody
			}
		} else {
			webhook.Headers = make(http.Header)
		}

		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, coreHttp.ErrInternal
	}

	return webhooks, nil
}
