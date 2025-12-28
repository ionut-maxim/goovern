package importer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ionut-maxim/goovern/ckan"
)

type DownloadArgs struct {
	Resource ckan.Resource `json:"resourceGetter"`
}

func (args DownloadArgs) Kind() string {
	return "download"
}

type DownloadWorker struct {
	store  ResourceStore
	jobs   *river.Client[pgx.Tx]
	logger *slog.Logger

	river.WorkerDefaults[DownloadArgs]
}

func NewDownloadWorker(jobs *river.Client[pgx.Tx], store ResourceStore, logger *slog.Logger) (*DownloadWorker, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if store == nil {
		return nil, errors.New("store required")
	}

	return &DownloadWorker{
		store:  store,
		jobs:   jobs,
		logger: logger.With("worker", "download"),
	}, nil
}

func (w *DownloadWorker) Timeout(*river.Job[DownloadArgs]) time.Duration { return 60 * time.Minute }

// NextRetry configures retry behavior for failed downloads
func (w *DownloadWorker) NextRetry(job *river.Job[DownloadArgs]) time.Time {
	// Retry with exponential backoff: 1min, 5min, 15min, 30min, 1hr
	// River will stop retrying after max attempts (default: 25)
	retryIntervals := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
	}

	attempt := job.Attempt
	if attempt >= len(retryIntervals) {
		// After 5 attempts, keep retrying every hour
		return time.Now().Add(1 * time.Hour)
	}

	return time.Now().Add(retryIntervals[attempt])
}

func (w *DownloadWorker) Work(ctx context.Context, job *river.Job[DownloadArgs]) error {
	resource := job.Args.Resource
	startTime := time.Now()

	logger := w.logger.With(
		"resource_id", resource.Id,
		"resource_name", resource.Name,
		"attempt", job.Attempt,
	)

	logger.Info("Starting download", "url", resource.Url, "priority", job.Priority)

	if err := w.store.Save(ctx, resource); err != nil {
		logger.Error("Download failed", "error", err)
		return err
	}

	duration := time.Since(startTime).Seconds()
	logger.Info("Download completed successfully", "duration_seconds", duration)

	return nil
}
