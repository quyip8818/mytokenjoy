package workers

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// IngestWorker processes individual consume-log ingest jobs via River.
// River handles retry scheduling and dead-lettering automatically.
type IngestWorker struct {
	river.WorkerDefaults[jobs.IngestArgs]
	ingest domainusage.Ingestor
	logger *slog.Logger
}

func NewIngestWorker(ingest domainusage.Ingestor, logger *slog.Logger) *IngestWorker {
	return &IngestWorker{ingest: ingest, logger: logger}
}

func (w *IngestWorker) Work(ctx context.Context, job *river.Job[jobs.IngestArgs]) error {
	source := job.Args.Source
	if source == "" {
		source = "webhook"
	}

	err := w.ingest.IngestByLogID(ctx, job.Args.LogID, source)
	if err == nil {
		return nil
	}

	kind := domainusage.ClassifyIngestError(err)
	switch kind {
	case domainusage.IngestBusiness:
		// Permanent business error (bad data, unprocessable) — don't retry.
		if !domainusage.IsRecoverableIngestError(err) {
			w.logger.Warn("ingest job cancelled (permanent business error)",
				"log_id", job.Args.LogID,
				"attempt", job.Attempt,
				"error", err,
			)
			return river.JobCancel(err)
		}
		// Recoverable (e.g. mapping not found yet) — let River retry.
		return err
	case domainusage.IngestLogNotFound:
		// Log not found — may appear later (replication lag). Let River retry.
		return err
	default:
		// IngestLogDBTemp, IngestRetryable — let River retry with its backoff.
		return err
	}
}
