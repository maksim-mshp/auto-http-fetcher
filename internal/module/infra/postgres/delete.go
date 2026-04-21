package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"

	"context"
)

func (r *PGModuleRepo) DeleteModule(ctx context.Context, moduleID, userID int) error {
	query := `DELETE FROM modules WHERE id = $1 AND owner_id = $2`

	cmdTag, err := r.pool.Exec(ctx, query, moduleID, userID)
	if err != nil {
		return coreHttp.ErrInternal
	}

	if cmdTag.RowsAffected() == 0 {
		return nil
	}

	return nil
}
