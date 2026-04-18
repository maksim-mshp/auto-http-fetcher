package infra

import (
	"auto-http-fetcher/internal/response/domain"
	"context"
	"encoding/json"
	"time"
)

func (pg *PGResponseRepo) FindByWebhookID(ctx context.Context, webhookID string) ([]*domain.Response, error) {
	var responses []*domain.Response
	query := `SELECT * FROM responses WHERE webhook_id = $1`

	rows, err := pg.pool.Query(ctx, query, webhookID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var response domain.Response
		var headers []byte
		var duration int64

		err := rows.Scan(
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

		if err = json.Unmarshal(headers, &response.Headers); err != nil {
			return nil, err
		}

		response.Duration = time.Duration(duration) * time.Millisecond

		responses = append(responses, &response)
	}

	return responses, rows.Err()
}
