package store

import (
	"github.com/felip/api-fidelidade/internal/store/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}
