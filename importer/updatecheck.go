package importer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/ionut-maxim/goovern/ckan"
	"github.com/ionut-maxim/goovern/db"
)

var packages = []string{}

type UpdateCheckArgs struct{}

func (u UpdateCheckArgs) Kind() string {
	return "updates"
}

type resourceGetter interface {
	Resource(ctx context.Context, tx db.Tx, id uuid.UUID) (ckan.Resource, bool, error)
}

type UpdatesWorker struct {
	ckanClient *ckan.Client
	jobs       *river.Client[pgx.Tx]
	logger     *slog.Logger
	db         db.Tx
	store      ResourceStore
	resource   resourceGetter

	river.WorkerDefaults[UpdateCheckArgs]
}

func NewUpdatesWorker(jobs *river.Client[pgx.Tx], db db.Tx, res resourceGetter, logger *slog.Logger) (*UpdatesWorker, error) {
	client, err := ckan.New()
	if err != nil {
		return nil, err
	}

	// Prevent panics
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	if db == nil {
		return nil, errors.New("db required")
	}

	if res == nil {
		return nil, errors.New("resourceGetter required")
	}

	return &UpdatesWorker{
		ckanClient: client,
		jobs:       jobs,
		db:         db,
		resource:   res,
		logger:     logger.With("worker", "updates"),
	}, nil
}

func (w *UpdatesWorker) Timeout(*river.Job[UpdateCheckArgs]) time.Duration {
	// Allow up to 6 hours for the entire update cycle (downloads + imports)
	return 6 * time.Hour
}

//CSV Import Order
//Tier 1 - No dependencies (import first):
//
//N_VERSIUNE_CAEN.CSV → caen_versions table
//N_STARE_FIRMA.CSV → company_statuses table
//Tier 2 - Depends on Tier 1:
//3. N_CAEN.CSV → caen_codes table (requires caen_versions)
//4. OD_FIRME.CSV → companies table (no dependencies)
//
//Tier 3 - Depends on Tier 2:
//5. OD_CAEN_AUTORIZAT.CSV → authorized_activities table (requires companies + caen_versions)
//6. OD_STARE_FIRMA.CSV → company_status_history table (requires companies + company_statuses)
//7. OD_REPREZENTANTI_LEGALI.CSV → legal_representatives table (requires companies)
//8. OD_REPREZENTANTI_IF.CSV → family_business_representatives table (requires companies)
//9. OD_SUCURSALE_ALTE_STATE_MEMBRE.CSV → foreign_branches table (requires companies)

func (w *UpdatesWorker) Work(ctx context.Context, _ *river.Job[UpdateCheckArgs]) error {
	startTime := time.Now()

	logger := w.logger

	logger.Info("Starting update check")

	s, err := w.ckanClient.Search(ctx, "onrc", 2)
	if err != nil {
		logger.Error("Failed to search CKAN", "error", err)
		return err
	}
	logger.Info("CKAN search completed", "packages_found", len(s.Results))

	// Collect all new resources
	var newResources []ckan.Resource
	for _, p := range s.Results {
		logger.Debug("Processing package", "package_name", p.Name, "resources_count", len(p.Resources))

		for _, resource := range p.Resources {
			_, exists, err := w.resource.Resource(ctx, w.db, resource.Id)
			if err != nil {
				logger.Error("Failed to check resource existence", "resource_id", resource.Id, "error", err)
				return err
			}
			if exists {
				logger.Debug("Resource already exists", "resource_id", resource.Id, "resource_name", resource.Name)
				continue
			}

			logger.Info("Found new resource", "resource_id", resource.Id, "resource_name", resource.Name)
			newResources = append(newResources, resource)
		}
	}

	if len(newResources) == 0 {
		duration := time.Since(startTime).Seconds()
		logger.Info("Update check complete - no new resources found", "duration_seconds", duration)
		return nil
	}

	logger.Info("Starting download phase", "total_resources", len(newResources))

	// Schedule all downloads in parallel
	tx, err := w.db.Begin(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	var downloadJobs []river.InsertManyParams
	for _, resource := range newResources {
		downloadJobs = append(downloadJobs, river.InsertManyParams{
			Args: DownloadArgs{Resource: resource},
			// No priority needed for downloads - they can all run in parallel
			InsertOpts: &river.InsertOpts{},
		})
	}

	downloadResults, err := w.jobs.InsertManyTx(ctx, tx, downloadJobs)
	if err != nil {
		logger.Error("Failed to insert download jobs", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	logger.Info("Download jobs scheduled", "count", len(downloadResults))

	// Wait for all downloads to complete
	logger.Info("Waiting for all downloads to complete")
	if err = w.waitForJobs(ctx, downloadResults); err != nil {
		logger.Error("Downloads failed", "error", err)
		return err
	}

	logger.Info("All downloads completed successfully")

	// Now schedule imports in the correct order
	logger.Info("Starting import phase")
	if err = w.scheduleImports(ctx, newResources); err != nil {
		logger.Error("Failed to schedule imports", "error", err)
		return err
	}

	duration := time.Since(startTime).Seconds()
	logger.Info("Update check complete", "resources_processed", len(newResources), "duration_seconds", duration)

	return nil
}

// waitForJobs polls River until all jobs are completed or failed
func (w *UpdatesWorker) waitForJobs(ctx context.Context, jobs []*rivertype.JobInsertResult) error {
	jobIDs := make(map[int64]bool)
	for _, job := range jobs {
		jobIDs[job.Job.ID] = true
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			allComplete := true
			for jobID := range jobIDs {
				job, err := w.jobs.JobGet(ctx, jobID)
				if err != nil {
					w.logger.Error("Failed to get job status", "job_id", jobID, "error", err)
					return err
				}

				switch job.State {
				case rivertype.JobStateCompleted:
					delete(jobIDs, jobID)
					w.logger.Debug("Job completed", "job_id", jobID)
				case rivertype.JobStateDiscarded, rivertype.JobStateCancelled:
					w.logger.Error("Job failed", "job_id", jobID, "state", job.State)
					return errors.New("job failed")
				default:
					allComplete = false
				}
			}

			if allComplete {
				return nil
			}

			w.logger.Debug("Waiting for jobs to complete", "remaining", len(jobIDs))
		}
	}
}

// scheduleImports schedules import jobs in the correct order based on dependencies
func (w *UpdatesWorker) scheduleImports(ctx context.Context, resources []ckan.Resource) error {
	// Group resources by priority tier
	var tier1 []ckan.Resource // N_VERSIUNE_CAEN.CSV, N_STARE_FIRMA.CSV
	var tier2 []ckan.Resource // N_CAEN.CSV
	var tier3 []ckan.Resource // OD_FIRME.CSV
	var tier4 []ckan.Resource // Everything else

	for _, resource := range resources {
		switch resource.Name {
		case "N_VERSIUNE_CAEN.CSV", "N_STARE_FIRMA.CSV":
			tier1 = append(tier1, resource)
		case "N_CAEN.CSV":
			tier2 = append(tier2, resource)
		case "OD_FIRME.CSV":
			tier3 = append(tier3, resource)
		default:
			tier4 = append(tier4, resource)
		}
	}

	// Schedule and wait for each tier to complete before moving to next
	tiers := []struct {
		name      string
		resources []ckan.Resource
	}{
		{"Tier 1 (caen_versions, company_statuses)", tier1},
		{"Tier 2 (caen_codes)", tier2},
		{"Tier 3 (companies)", tier3},
		{"Tier 4 (dependent tables)", tier4},
	}

	for _, tier := range tiers {
		if len(tier.resources) == 0 {
			w.logger.Debug("Skipping empty tier", "tier", tier.name)
			continue
		}

		w.logger.Info("Starting import tier", "tier", tier.name, "count", len(tier.resources))

		tx, err := w.db.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		var importJobs []river.InsertManyParams
		for _, resource := range tier.resources {
			importJobs = append(importJobs, river.InsertManyParams{
				Args:       ImportArgs{Resource: resource},
				InsertOpts: &river.InsertOpts{},
			})
		}

		results, err := w.jobs.InsertManyTx(ctx, tx, importJobs)
		if err != nil {
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			return err
		}

		w.logger.Info("Import jobs scheduled", "tier", tier.name, "count", len(results))

		// Wait for this tier to complete before moving to next
		if err = w.waitForJobs(ctx, results); err != nil {
			return err
		}

		w.logger.Info("Import tier completed", "tier", tier.name)
	}

	return nil
}
