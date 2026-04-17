package postgres

import (
	httpCore "auto-http-fetcher/internal/core/http"
	"context"
)

func (ur *PGUserRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := ur.pool.Exec(ctx, query, id)
	if err != nil {
		return httpCore.ErrInternal
	}

	if cmdTag.RowsAffected() == 0 {
		return httpCore.ErrUserNotFound
	}

	return nil
}
