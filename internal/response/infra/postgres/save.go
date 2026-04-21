package infra

import (
	"auto-http-fetcher/internal/response/domain"
	"context"
	"encoding/json"
)

func (pg *PGResponseRepo) Save(ctx context.Context, r *domain.Response) error {
	query := `INSERT INTO responses (webhook_id, type, status, status_code, body, headers, started_at, finished_at, attempt, duration)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	headers, err := json.Marshal(r.Headers)
	if err != nil {
		return err
	}

	err = pg.pool.QueryRow(ctx, query,
		r.WebhookID,
		r.Type,
		r.Status,
		r.StatusCode,
		r.Body,
		headers,
		r.StartedAt,
		r.FinishedAt,
		r.Attempt,
		r.Duration.Milliseconds(),
	).Scan(&r.ID)

	return err
}
