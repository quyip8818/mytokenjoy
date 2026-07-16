package workers

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

// IngestReconcileWorker scans consume logs after the stored cursor and enqueues
// any that haven't been ingested yet. This replaces the old reconcile loop.
type IngestReconcileWorker struct {
	river.WorkerDefaults[jobs.IngestReconcileArgs]
	logStore  store.LogStore
	ingest    domainusage.Ingestor
	enqueuer  jobs.Enqueuer
	batchSize int
	maxRounds int
	logger    *slog.Logger
}

func NewIngestReconcileWorker(
	logStore store.LogStore,
	ingest domainusage.Ingestor,
	enqueuer jobs.Enqueuer,
	batchSize int,
	maxRounds int,
	logger *slog.Logger,
) *IngestReconcileWorker {
	if batchSize <= 0 {
		batchSize = 500
	}
	if maxRounds <= 0 {
		maxRounds = 10
	}
	return &IngestReconcileWorker{
		logStore:  logStore,
		ingest:    ingest,
		enqueuer:  enqueuer,
		batchSize: batchSize,
		maxRounds: maxRounds,
		logger:    logger,
	}
}

func (w *IngestReconcileWorker) Work(ctx context.Context, _ *river.Job[jobs.IngestReconcileArgs]) error {
	for round := 0; round < w.maxRounds; round++ {
		cursor, err := w.logStore.GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
		if err != nil {
			return err
		}
		ids, err := w.logStore.ListConsumeLogIDsAfter(ctx, cursor, w.batchSize)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, id := range ids {
			// Enqueue each missing log as a River ingest job (idempotent via UniqueOpts).
			if err := jobs.InsertIngest(ctx, w.enqueuer, id, "reconcile"); err != nil {
				w.logger.Warn("reconcile: enqueue ingest job failed", "log_id", id, "error", err)
			}
		}
		// Advance cursor once per batch to the last processed ID.
		lastID := ids[len(ids)-1]
		if err := w.logStore.SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, lastID); err != nil {
			return err
		}
		if len(ids) < w.batchSize {
			return nil
		}
	}
	return nil
}
