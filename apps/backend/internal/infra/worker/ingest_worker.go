package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/store"
)

const ingestReconcileLockName = "ingest_reconcile"

type IngestWorker struct {
	cfg            config.Config
	logStore       store.LogStore
	ingest         domainusage.Ingestor
	queue          domainusage.Queue
	metrics        ingestmetrics.Recorder
	schedulerLock  store.SchedulerLockRepository
	logger         *slog.Logger
	holderID       string
	reconcileEvery time.Duration
}

func NewIngestWorker(
	cfg config.Config,
	logStore store.LogStore,
	ingest domainusage.Ingestor,
	queue domainusage.Queue,
	metrics ingestmetrics.Recorder,
	schedulerLock store.SchedulerLockRepository,
	logger *slog.Logger,
	holderID string,
	reconcileEvery time.Duration,
) *IngestWorker {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	if metrics == nil {
		metrics = ingestmetrics.NoopCollector()
	}
	if queue == nil {
		queue = domainusage.NewQueue(logStore)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &IngestWorker{
		cfg:            cfg,
		logStore:       logStore,
		ingest:         ingest,
		queue:          queue,
		metrics:        metrics,
		schedulerLock:  schedulerLock,
		logger:         logger,
		holderID:       holderID,
		reconcileEvery: reconcileEvery,
	}
}

func (w *IngestWorker) ProcessPending(ctx context.Context) error {
	if !w.cfg.IngestEnabled() {
		return nil
	}
	jobs, err := w.logStore.ClaimPendingJobs(ctx, w.cfg.JobBatchSize())
	if err != nil {
		return err
	}
	for _, job := range jobs {
		source := job.Source
		if source == "" {
			source = types.SourceWebhook
		}
		ingestErr := w.ingest.IngestByLogID(ctx, job.LogID, source)
		if handleErr := w.queue.ApplyRetry(ctx, job, ingestErr); handleErr != nil {
			return handleErr
		}
	}
	w.refreshMetrics(ctx)
	return nil
}

func (w *IngestWorker) ProcessReconcile(ctx context.Context) error {
	if !w.cfg.IngestEnabled() {
		return nil
	}
	lease := w.reconcileEvery + time.Minute
	if lease < 2*time.Minute {
		lease = 2 * time.Minute
	}
	acquired, err := w.schedulerLock.TryAcquire(ctx, ingestReconcileLockName, w.holderID, lease)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() {
		_ = w.schedulerLock.Release(ctx, ingestReconcileLockName, w.holderID)
	}()

	batchSize := w.cfg.ReconcileBatchSize()
	maxRounds := w.cfg.ReconcileMaxRounds()
	for round := 0; round < maxRounds; round++ {
		cursor, err := w.logStore.GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
		if err != nil {
			return err
		}
		ids, err := w.logStore.ListConsumeLogIDsAfter(ctx, cursor, batchSize)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			w.refreshMetrics(ctx)
			return nil
		}
		for _, id := range ids {
			ingestErr := w.ingest.IngestByLogID(ctx, id, types.SourceReconcile)
			outcome := domainusage.OutcomeFor(ingestErr)
			if !outcome.ReconcileAdvancesCursor() {
				return ingestErr
			}
			if outcome.ShouldRecordFailure() {
				if recordErr := w.queue.RecordFailure(ctx, id, types.SourceReconcile, ingestErr); recordErr != nil {
					w.logger.Warn("upsert ingest failure", "log_id", id, "error", recordErr)
				}
			}
			if setErr := w.logStore.SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, id); setErr != nil {
				return setErr
			}
		}
		if len(ids) < batchSize {
			w.refreshMetrics(ctx)
			return nil
		}
	}
	w.refreshMetrics(ctx)
	return nil
}

func (w *IngestWorker) refreshMetrics(ctx context.Context) {
	if !w.cfg.IngestEnabled() {
		return
	}
	if err := w.metrics.Refresh(ctx, w.logStore); err != nil {
		w.logger.Warn("refresh ingest metrics failed", "error", err)
	}
}
