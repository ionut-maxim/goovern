package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivermigrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	// Configure goose to use embedded migrations from db package
	goose.SetBaseFS(migrations)

	// Open a standard database/sql connection for goose
	// (goose doesn't work directly with pgxpool)
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	migrator, err := rivermigrate.New(riverdatabasesql.New(sqlDB), &rivermigrate.Config{Logger: logger})
	if err != nil {
		return err
	}

	// Migrate up to the latest version supported by this River version
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		return err
	}

	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err = goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("migrations completed successfully")
	return nil
}
