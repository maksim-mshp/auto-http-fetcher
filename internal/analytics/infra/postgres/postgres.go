package infra

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGAnalyticsRepo struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepo(pool *pgxpool.Pool) *PGAnalyticsRepo {
	return &PGAnalyticsRepo{
		pool: pool,
	}
}
