package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type PGUserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *PGUserRepo {
	return &PGUserRepo{pool: pool}
}
