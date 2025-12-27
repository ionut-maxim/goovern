package importer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/ionut-maxim/goovern/ckan"
	"github.com/ionut-maxim/goovern/telemetry"
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

	// Metrics
	downloadDuration metric.Float64Histogram
	downloadsTotal   metric.Int64Counter
	bytesDownloaded  metric.Int64Counter

	river.WorkerDefaults[DownloadArgs]
}

func NewDownloadWorker(jobs *river.Client[pgx.Tx], store ResourceStore, logger *slog.Logger) (*DownloadWorker, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if store == nil {
		return nil, errors.New("store required")
	}

	// Initialize metrics
	meter := telemetry.Meter()
	downloadDuration, err := meter.Float64Histogram(
		"goovern.download.duration",
		metric.WithDescription("Duration of download operations in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	downloadsTotal, err := meter.Int64Counter(
		"goovern.download.total",
		metric.WithDescription("Total number of download operations"),
	)
	if err != nil {
		return nil, err
	}

	bytesDownloaded, err := meter.Int64Counter(
		"goovern.download.bytes",
		metric.WithDescription("Total bytes downloaded"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	return &DownloadWorker{
		store:            store,
		jobs:             jobs,
		logger:           logger.With("worker", "download"),
		downloadDuration: downloadDuration,
		downloadsTotal:   downloadsTotal,
		bytesDownloaded:  bytesDownloaded,
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

	ctx, span := telemetry.StartSpan(ctx, "download.work",
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

	// Metric attributes
	metricAttrs := metric.WithAttributes(
		attribute.String("resource.name", resource.Name),
		attribute.String("status", "success"),
	)

	logger.Info("Starting download", "url", resource.Url, "priority", job.Priority)

	if err := w.store.Save(ctx, resource); err != nil {
		logger.Error("Download failed", "error", err)
		telemetry.RecordError(span, err)

		// Record failure metrics
		duration := time.Since(startTime).Seconds()
		failureAttrs := metric.WithAttributes(
			attribute.String("resource.name", resource.Name),
			attribute.String("status", "failure"),
		)
		w.downloadDuration.Record(ctx, duration, failureAttrs)
		w.downloadsTotal.Add(ctx, 1, failureAttrs)

		return err
	}

	duration := time.Since(startTime).Seconds()
	logger.Info("Download completed successfully", "duration_seconds", duration)

	// Record success metrics
	w.downloadDuration.Record(ctx, duration, metricAttrs)
	w.downloadsTotal.Add(ctx, 1, metricAttrs)
	// Note: bytes downloaded would need to be returned from store.Save() to track accurately

	return nil
}
