package infra

import (
	"auto-http-fetcher/internal/response/domain"
	"context"
	"encoding/json"
	"time"
)

func (pg *PGResponseRepo) FindByID(ctx context.Context, id int) (*domain.Response, error) {
	query := `SELECT id, webhook_id, type, status, status_code, body, headers, started_at, finished_at, attempt, duration FROM responses WHERE id = $1`

	var response domain.Response
	var headers []byte
	var duration int64

	err := pg.pool.QueryRow(ctx, query, id).Scan(
		&response.ID,
		&response.WebhookID,
		&response.Type,
		&response.Status,
		&response.StatusCode,
		&response.Body,
		&headers,
		&response.StartedAt,
		&response.FinishedAt,
		&response.Attempt,
		&duration,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(headers, &response.Headers); err != nil {
		return nil, err
	}
	response.Duration = time.Duration(duration) * time.Millisecond

	return &response, nil
}