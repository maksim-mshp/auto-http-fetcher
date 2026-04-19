package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	httpCore "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/user/domain"
)

func (ur *PGUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `INSERT INTO users (name, email, password) 
					VALUES ($1, $2, $3)
					RETURNING id, name, email;`

	var newUser domain.User
	err := ur.pool.QueryRow(ctx, query, user.Name, user.Email, user.Password).Scan(
		&newUser.ID, &newUser.Name, &newUser.Email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, httpCore.ErrUserAlreadyExists
			}
		}
		return nil, httpCore.ErrInternal
	}
	return &newUser, nil
}
