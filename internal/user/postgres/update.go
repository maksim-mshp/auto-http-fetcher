package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	httpCore "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/user/domain"
)

func (ur *PGUserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `UPDATE users 
              SET name = $1, email = $2 
              WHERE id = $3
              RETURNING id, name, email`

	var updatedUser domain.User
	err := ur.pool.QueryRow(ctx, query, user.Name, user.Email, user.ID).Scan(
		&updatedUser.ID, &updatedUser.Name, &updatedUser.Email)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, httpCore.ErrUserAlreadyExists
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpCore.ErrUserNotFound
		}
		return nil, httpCore.ErrInternal
	}

	return &updatedUser, nil
}
