package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type PGWebhookRepo struct {
	pool *pgxpool.Pool
}

func NewPGWebhookRepo(pool *pgxpool.Pool) *PGWebhookRepo {
	return &PGWebhookRepo{pool: pool}
}
