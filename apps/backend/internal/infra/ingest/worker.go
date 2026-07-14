package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/store"
)

const ingestReconcileLockName = "ingest_reconcile"

type Worker struct {
	cfg            config.Config
	logStore       store.LogStore
	ingest         domainusage.Ingestor
	queue          domainusage.Queue
	metrics        ingestmetrics.Recorder
	schedulerLock  store.SchedulerLockRepository
	billing        domainbilling.Service
	logger         *slog.Logger
	holderID       string
	pollInterval   time.Duration
	reconcileEvery time.Duration
}

func NewWorker(
	cfg config.Config,
	logStore store.LogStore,
	ingest domainusage.Ingestor,
	queue domainusage.Queue,
	metrics ingestmetrics.Recorder,
	schedulerLock store.SchedulerLockRepository,
	billing domainbilling.Service,
	logger *slog.Logger,
) *Worker {
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
	return &Worker{
		cfg:            cfg,
		logStore:       logStore,
		ingest:         ingest,
		queue:          queue,
		metrics:        metrics,
		schedulerLock:  schedulerLock,
		billing:        billing,
		logger:         logger,
		holderID:       fmt.Sprintf("ingest-%d", time.Now().UnixNano()),
		pollInterval:   cfg.WorkerPollInterval(),
		reconcileEvery: cfg.IngestReconcileInterval(),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if !w.cfg.IngestEnabled() {
		w.logger.Warn("ingest worker not started: LOG_DATABASE_URL empty")
		return
	}
	w.logger.Info("ingest worker started",
		"poll_interval", w.pollInterval.String(),
		"reconcile_every", w.reconcileEvery.String(),
		"job_batch_size", w.cfg.JobBatchSize(),
	)
	go w.loop(ctx)
}

func (w *Worker) loop(ctx context.Context) {
	w.logStep("ingest_reconcile_startup", w.ProcessReconcile(ctx))
	ticker := time.NewTicker(w.pollInterval)
	reconcileTicker := time.NewTicker(w.reconcileEvery)
	defer ticker.Stop()
	defer reconcileTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.logStep("ingest_pending", w.ProcessPending(ctx))
			w.logStep("wallet_reconcile", w.processWalletReconcile(ctx))
		case <-reconcileTicker.C:
			w.logStep("ingest_reconcile", w.ProcessReconcile(ctx))
		}
	}
}

func (w *Worker) ProcessPending(ctx context.Context) error {
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
		if ingestErr != nil {
			disposition := domainusage.OutcomeFor(ingestErr).Retry(job.Attempts)
			w.logger.Warn("ingest job failed",
				"log_id", job.LogID,
				"job_id", job.ID,
				"source", source,
				"attempts", job.Attempts,
				"disposition", disposition.String(),
				"error", ingestErr,
			)
		}
		if handleErr := w.queue.ApplyRetry(ctx, job, ingestErr); handleErr != nil {
			return handleErr
		}
	}
	w.refreshMetrics(ctx)
	return nil
}

func (w *Worker) ProcessReconcile(ctx context.Context) error {
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

func (w *Worker) processWalletReconcile(ctx context.Context) error {
	if w.billing == nil {
		return nil
	}
	return w.billing.ReconcileWalletDrift(ctx)
}

func (w *Worker) refreshMetrics(ctx context.Context) {
	if !w.cfg.IngestEnabled() {
		return
	}
	if err := w.metrics.Refresh(ctx, w.logStore); err != nil {
		w.logger.Warn("refresh ingest metrics failed", "error", err)
	}
}

func (w *Worker) logStep(step string, err error) {
	if err != nil {
		w.logger.Warn("ingest worker step failed", "step", step, "error", err)
	}
}

// RunReconcileOnce runs a single reconcile pass (tests).
func (w *Worker) RunReconcileOnce(ctx context.Context) error {
	return w.ProcessReconcile(ctx)
}

// RunPendingOnce processes one pending batch (tests).
func (w *Worker) RunPendingOnce(ctx context.Context) error {
	return w.ProcessPending(ctx)
}
