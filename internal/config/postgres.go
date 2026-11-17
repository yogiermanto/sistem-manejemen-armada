package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabase(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
