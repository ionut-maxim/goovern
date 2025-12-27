package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ionut-maxim/goovern/db"
)

func newDB(databaseURL string, logger *slog.Logger) (*pgxpool.Pool, *db.DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pool: %v", err)
	}

	migrateCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	if err = db.Migrate(migrateCtx, pool, logger); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate: %v", err)
	}

	return pool, db.New(logger), nil
}
