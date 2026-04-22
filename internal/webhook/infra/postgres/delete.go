package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"

	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *PGWebhookRepo) DeleteWebhook(ctx context.Context, webhookID, moduleID, userID int) error {
	checkQuery := `SELECT EXISTS(SELECT 1 FROM modules WHERE id = $1 AND owner_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, checkQuery, moduleID, userID).Scan(&exists)
	if err != nil {
		return coreHttp.ErrInternal
	}
	if !exists {
		return coreHttp.ErrModuleNotFound
	}

	query := `DELETE FROM webhooks WHERE id = $1 AND module_id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, webhookID, moduleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return coreHttp.ErrWebhookInUse
		}
		return coreHttp.ErrInternal
	}

	if cmdTag.RowsAffected() == 0 {
		return coreHttp.ErrWebhookNotFound
	}

	return nil
}
