package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type PGModuleRepo struct {
	pool *pgxpool.Pool
}

func NewPGModuleRepo(pool *pgxpool.Pool) *PGModuleRepo {
	return &PGModuleRepo{pool: pool}
}
