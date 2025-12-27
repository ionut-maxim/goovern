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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/ionut-maxim/goovern/ckan"
	"github.com/ionut-maxim/goovern/db"
	"github.com/ionut-maxim/goovern/telemetry"
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

	// Metrics
	importDuration metric.Float64Histogram
	importsTotal   metric.Int64Counter
	rowsImported   metric.Int64Counter

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

	// Initialize metrics
	meter := telemetry.Meter()
	importDuration, err := meter.Float64Histogram(
		"goovern.import.duration",
		metric.WithDescription("Duration of import operations in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	importsTotal, err := meter.Int64Counter(
		"goovern.import.total",
		metric.WithDescription("Total number of import operations"),
	)
	if err != nil {
		return nil, err
	}

	rowsImported, err := meter.Int64Counter(
		"goovern.import.rows",
		metric.WithDescription("Total number of rows imported"),
	)
	if err != nil {
		return nil, err
	}

	return &ImportWorker{
		db:             pool,
		repo:           repo,
		store:          store,
		logger:         logger.With("worker", "import"),
		importDuration: importDuration,
		importsTotal:   importsTotal,
		rowsImported:   rowsImported,
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

	// Start span for tracing
	ctx, span := telemetry.StartSpan(ctx, "import.work",
		attribute.String("resource.id", resource.Id.String()),
		attribute.String("resource.name", resource.Name),
		attribute.Int("job.attempt", job.Attempt),
		attribute.Int("job.priority", job.Priority),
	)
	defer span.End()

	// Create logger with trace context for correlation
	logger := telemetry.LoggerWithTrace(ctx, w.logger).With(
		"resource_id", resource.Id,
		"resource_name", resource.Name,
		"attempt", job.Attempt,
	)

	logger.Info("Starting import", "priority", job.Priority)

	// Metric attributes for this import
	metricAttrs := metric.WithAttributes(
		attribute.String("resource.name", resource.Name),
		attribute.String("status", "success"),
	)

	tx, err := w.db.Begin(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		telemetry.RecordError(span, err)
		w.recordFailure(ctx, resource, startTime)
		return err
	}
	defer tx.Rollback(ctx)

	logger.Debug("Loading file from store")
	data, err := w.store.Load(ctx, resource)
	if err != nil {
		logger.Error("Failed to load file from store", "error", err)
		telemetry.RecordError(span, err)
		w.recordFailure(ctx, resource, startTime)
		return err
	}
	defer data.Close()

	logger.Info("Importing data to database")
	if err = w.repo.Import(ctx, tx, resource, data); err != nil {
		logger.Error("Import failed", "error", err)
		telemetry.RecordError(span, err)
		w.recordFailure(ctx, resource, startTime)
		return err
	}

	logger.Debug("Saving resource metadata")
	if err = w.repo.SaveResource(ctx, tx, resource); err != nil {
		logger.Error("Failed to save resource metadata", "error", err)
		telemetry.RecordError(span, err)
		w.recordFailure(ctx, resource, startTime)
		return err
	}

	logger.Debug("Committing transaction")
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		telemetry.RecordError(span, err)
		w.recordFailure(ctx, resource, startTime)
		return err
	}

	duration := time.Since(startTime).Seconds()
	logger.Info("Import completed successfully", "duration_seconds", duration)

	// Record success metrics
	w.importDuration.Record(ctx, duration, metricAttrs)
	w.importsTotal.Add(ctx, 1, metricAttrs)

	return nil
}

func (w *ImportWorker) recordFailure(ctx context.Context, resource ckan.Resource, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	metricAttrs := metric.WithAttributes(
		attribute.String("resource.name", resource.Name),
		attribute.String("status", "failure"),
	)
	w.importDuration.Record(ctx, duration, metricAttrs)
	w.importsTotal.Add(ctx, 1, metricAttrs)
}
