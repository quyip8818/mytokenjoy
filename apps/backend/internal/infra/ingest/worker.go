package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/store"
)

const (
	ingestReconcileLockName    = "ingest_reconcile"
	ingestListenFallbackPoll   = 5 * time.Second
	ingestCompanyGroupPoolSize = 8
)

type companyResolver interface {
	CompanyIDsByLogID(ctx context.Context, logIDs []int64) (map[int64]int64, error)
}

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
	listener       store.PGListener
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

// SetListener attaches a PGListener for LISTEN/NOTIFY driven processing.
// When set, the worker wakes immediately on notifications instead of only polling.
func (w *Worker) SetListener(l store.PGListener) {
	w.listener = l
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

	notify := make(chan struct{}, 1)
	listenOK := false
	if w.listener != nil {
		if err := w.listener.Listen(ctx, store.IngestPendingChannel); err != nil {
			w.logger.Warn("LISTEN setup failed, falling back to poll only", "error", err)
		} else {
			listenOK = true
			go func() {
				for {
					if err := w.listener.WaitForNotification(ctx); err != nil {
						if ctx.Err() != nil {
							return
						}
						w.logger.Warn("LISTEN notification error", "error", err)
						time.Sleep(time.Second)
						continue
					}
					select {
					case notify <- struct{}{}:
					default:
					}
				}
			}()
		}
		defer func() { _ = w.listener.Close(context.Background()) }()
	}

	pollEvery := w.pollInterval
	if listenOK {
		pollEvery = ingestListenFallbackPoll
	}
	w.logger.Info("ingest worker poll configured",
		"poll_every", pollEvery.String(),
		"listen_enabled", listenOK,
	)

	pollTicker := time.NewTicker(pollEvery)
	reconcileTicker := time.NewTicker(w.reconcileEvery)
	defer pollTicker.Stop()
	defer reconcileTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-notify:
			w.logStep("ingest_pending", w.ProcessPending(ctx))
		case <-pollTicker.C:
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
	if len(jobs) == 0 {
		return nil
	}

	groups := w.pendingJobGroups(ctx, jobs)

	type jobResult struct {
		job       store.IngestJob
		ingestErr error
	}
	results := make(chan jobResult, len(jobs))
	sem := make(chan struct{}, ingestCompanyGroupPoolSize)
	var wg sync.WaitGroup

	for _, group := range groups {
		sem <- struct{}{}
		wg.Add(1)
		go func(batch []store.IngestJob) {
			defer func() { <-sem; wg.Done() }()
			for _, j := range batch {
				source := j.Source
				if source == "" {
					source = types.SourceWebhook
				}
				ingestErr := w.ingest.IngestByLogID(ctx, j.LogID, source)
				if ingestErr != nil {
					disposition := domainusage.OutcomeFor(ingestErr).Retry(j.Attempts)
					w.logger.Warn("ingest job failed",
						"log_id", j.LogID,
						"job_id", j.ID,
						"source", source,
						"attempts", j.Attempts,
						"disposition", disposition.String(),
						"error", ingestErr,
					)
				}
				results <- jobResult{job: j, ingestErr: ingestErr}
			}
		}(group)
	}

	wg.Wait()
	close(results)

	for r := range results {
		if handleErr := w.queue.ApplyRetry(ctx, r.job, r.ingestErr); handleErr != nil {
			return handleErr
		}
	}
	w.refreshMetrics(ctx)
	return nil
}

func (w *Worker) pendingJobGroups(ctx context.Context, jobs []store.IngestJob) [][]store.IngestJob {
	resolver, ok := w.ingest.(companyResolver)
	if !ok {
		return [][]store.IngestJob{jobs}
	}
	logIDs := make([]int64, len(jobs))
	for i, job := range jobs {
		logIDs[i] = job.LogID
	}
	companyByLogID, err := resolver.CompanyIDsByLogID(ctx, logIDs)
	if err != nil {
		w.logger.Warn("resolve ingest job companies failed; processing serial within batch", "error", err)
		return [][]store.IngestJob{jobs}
	}
	return groupJobsByCompany(jobs, companyByLogID)
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
