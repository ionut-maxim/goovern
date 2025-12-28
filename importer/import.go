package importer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/ionut-maxim/goovern/ckan"
	"github.com/ionut-maxim/goovern/db"
)

type ImportArgs struct {
	Resource ckan.Resource `json:"resourceGetter"`
}

func (i ImportArgs) Kind() string {
	return "import"
}

type repo interface {
	SaveResource(ctx context.Context, tx db.Tx, resource ckan.Resource) error
	Import(ctx context.Context, db db.Tx, resource ckan.Resource, data io.Reader) error
}

type ImportWorker struct {
	db     db.Tx
	repo   repo
	store  ResourceStore
	logger *slog.Logger

	river.WorkerDefaults[ImportArgs]
}

func NewImportWorker(pool *pgxpool.Pool, store ResourceStore, repo repo, logger *slog.Logger) (*ImportWorker, error) {
	if pool == nil {
		return nil, errors.New("db required")
	}
	if store == nil {
		return nil, errors.New("store required")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &ImportWorker{
		db:     pool,
		repo:   repo,
		store:  store,
		logger: logger.With("worker", "import"),
	}, nil
}

func (w *ImportWorker) Timeout(*river.Job[ImportArgs]) time.Duration { return 30 * time.Minute }

// NextRetry configures retry behavior for failed imports
func (w *ImportWorker) NextRetry(job *river.Job[ImportArgs]) time.Time {
	// Retry with backoff: 30s, 2min, 5min, 10min
	retryIntervals := []time.Duration{
		30 * time.Second,
		2 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
	}

	attempt := job.Attempt
	if attempt >= len(retryIntervals) {
		// After 4 attempts, keep retrying every 10 minutes
		return time.Now().Add(10 * time.Minute)
	}

	return time.Now().Add(retryIntervals[attempt])
}

func (w *ImportWorker) Work(ctx context.Context, job *river.Job[ImportArgs]) error {
	resource := job.Args.Resource
	startTime := time.Now()

	logger := w.logger.With(
		"resource_id", resource.Id,
		"resource_name", resource.Name,
		"attempt", job.Attempt,
	)

	logger.Info("Starting import", "priority", job.Priority)

	tx, err := w.db.Begin(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	logger.Debug("Loading file from store")
	data, err := w.store.Load(ctx, resource)
	if err != nil {
		logger.Error("Failed to load file from store", "error", err)
		return err
	}
	defer data.Close()

	logger.Info("Importing data to database")
	if err = w.repo.Import(ctx, tx, resource, data); err != nil {
		logger.Error("Import failed", "error", err)
		return err
	}

	logger.Debug("Saving resource metadata")
	if err = w.repo.SaveResource(ctx, tx, resource); err != nil {
		logger.Error("Failed to save resource metadata", "error", err)
		return err
	}

	logger.Debug("Committing transaction")
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	duration := time.Since(startTime).Seconds()
	logger.Info("Import completed successfully", "duration_seconds", duration)

	return nil
}
