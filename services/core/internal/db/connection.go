package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Dbc(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
