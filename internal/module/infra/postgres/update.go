package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	domainModule "auto-http-fetcher/internal/module/domain"

	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PGModuleRepo) UpdateModule(ctx context.Context, module domainModule.Module, userID int) (
	*domainModule.Module, error) {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	defer tx.Rollback(ctx)

	query := `UPDATE modules 
              SET name = $1, description = $2, updated_at = NOW() 
              WHERE id = $3 AND owner_id = $4
              RETURNING id`

	err = tx.QueryRow(ctx, query, module.Name, module.Description, module.ID, userID).
		Scan(&module.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, coreHttp.ErrModuleNotFound
		}
		return nil, coreHttp.ErrInternal
	}

	deleteQuery := `DELETE FROM webhooks WHERE module_id = $1`
	if _, err = tx.Exec(ctx, deleteQuery, module.ID); err != nil {
		return nil, coreHttp.ErrInternal
	}

	for _, webhook := range module.Webhooks {
		if err = r.createWebhook(ctx, tx, webhook); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, coreHttp.ErrInternal
	}

	return &module, nil
}
