package worker

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/robfig/cron/v3"

	"github.com/ionut-maxim/goovern/db"
	"github.com/ionut-maxim/goovern/importer"
)

func New(pool *pgxpool.Pool, db *db.DB, logger *slog.Logger) (*river.Client[pgx.Tx], error) {
	jobsClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		return nil, err
	}

	resourceStore, err := importer.NewFSResourceStore("data", logger)
	if err != nil {
		return nil, err
	}

	updatesWorker, err := importer.NewUpdatesWorker(jobsClient, pool, db, logger)
	if err != nil {
		return nil, err
	}

	downloadWorker, err := importer.NewDownloadWorker(jobsClient, resourceStore, logger)
	if err != nil {
		return nil, err
	}

	importWorker, err := importer.NewImportWorker(pool, resourceStore, db, logger)
	if err != nil {
		return nil, err
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, updatesWorker)
	river.AddWorker(workers, downloadWorker)
	river.AddWorker(workers, importWorker)

	schedule, err := cron.ParseStandard("@midnight")
	if err != nil {
		return nil, err
	}

	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			schedule,
			func() (river.JobArgs, *river.InsertOpts) {
				return importer.UpdateCheckArgs{}, &river.InsertOpts{}
			},
			&river.PeriodicJobOpts{RunOnStart: true, ID: "update-checker"},
		),
	}

	workClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5}, // Allow 5 parallel workers (for downloads)
		},
		Workers:                     workers,
		PeriodicJobs:                periodicJobs,
		Logger:                      logger,
		CancelledJobRetentionPeriod: 24 * time.Hour,
	})
	if err != nil {
		return nil, err
	}

	return workClient, nil
}
