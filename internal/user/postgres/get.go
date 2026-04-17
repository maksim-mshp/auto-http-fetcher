package postgres

import (
	httpCore "auto-http-fetcher/internal/core/http"
	"auto-http-fetcher/internal/user/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

func (ur *PGUserRepo) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE id = $1`

	var user domain.User
	err := ur.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpCore.ErrUserNotFound
		}
		return nil, httpCore.ErrInternal
	}

	return &user, nil
}

func (ur *PGUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE email = $1`

	var user domain.User
	err := ur.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpCore.ErrUserNotFound
		}
		return nil, httpCore.ErrInternal
	}

	return &user, nil
}
