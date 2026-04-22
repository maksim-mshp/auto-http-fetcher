package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"

	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *PGModuleRepo) DeleteModule(ctx context.Context, moduleID, userID int) error {
	query := `DELETE FROM modules WHERE id = $1 AND owner_id = $2`

	cmdTag, err := r.pool.Exec(ctx, query, moduleID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return coreHttp.ErrModuleHasWebhooks
		}
		return coreHttp.ErrInternal
	}

	if cmdTag.RowsAffected() == 0 {
		return coreHttp.ErrModuleNotFound
	}

	return nil
}
