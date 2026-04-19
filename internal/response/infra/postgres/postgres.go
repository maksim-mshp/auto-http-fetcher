package infra

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGResponseRepo struct {
	pool *pgxpool.Pool
}

func NewResponseRepo(pool *pgxpool.Pool) *PGResponseRepo {
	return &PGResponseRepo{pool: pool}
}
