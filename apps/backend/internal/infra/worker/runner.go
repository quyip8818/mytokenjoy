package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type Runner struct {
	cfg           config.Config
	relayJobs     store.RelayJobRepository
	schedulerLock store.SchedulerLockRepository
	relaySync     relay.RelayOutboxSync
	overrun       domainbudget.OverrunProcessor
	rebalance     domainbudget.Rebalancer
	walletSync    billingWalletSync
	syncSvc       domainorg.SyncService
	ingestWorker  *IngestWorker
	logger        *slog.Logger
	interval      time.Duration
	syncEvery     time.Duration
	syncTick      time.Duration

	lastRebalanceMonth string
}

func NewRunner(
	cfg config.Config,
	relayRepo store.RelayRepository,
	schedulerLock store.SchedulerLockRepository,
	logStore store.LogStore,
	metrics ingestmetrics.Recorder,
	relaySync relay.RelayOutboxSync,
	ingest domainusage.Ingestor,
	failureRecorder domainusage.FailureRecorder,
	overrun domainbudget.OverrunProcessor,
	rebalance domainbudget.Rebalancer,
	billingSvc domainbilling.Service,
	syncSvc domainorg.SyncService,
	logger *slog.Logger,
) *Runner {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	if metrics == nil {
		metrics = ingestmetrics.NoopCollector()
	}
	if failureRecorder == nil {
		failureRecorder = domainusage.NewFailureRecorder(logStore, logger)
	}
	holderID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	return &Runner{
		cfg:           cfg,
		relayJobs:     relayRepo,
		schedulerLock: schedulerLock,
		relaySync:     relaySync,
		overrun:       overrun,
		rebalance:     rebalance,
		walletSync:    billingWalletSync{svc: billingSvc},
		syncSvc:       syncSvc,
		ingestWorker: NewIngestWorker(
			cfg, logStore, ingest, metrics, schedulerLock, failureRecorder, logger, holderID, cfg.IngestReconcileInterval(),
		),
		logger:             logger,
		interval:           cfg.WorkerPollInterval(),
		syncEvery:          cfg.WorkerOrgSyncInterval(),
		lastRebalanceMonth: pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String(),
	}
}

func (r *Runner) Start(ctx context.Context) {
	if r.cfg.IngestEnabled() {
		go r.ingestLoop(ctx)
	}
	go r.relayLoop(ctx)
}

func (r *Runner) ingestLoop(ctx context.Context) {
	r.logStep("ingest_reconcile_startup", r.ingestWorker.ProcessReconcile(ctx))
	ticker := time.NewTicker(r.interval)
	reconcileTicker := time.NewTicker(r.ingestWorker.reconcileEvery)
	defer ticker.Stop()
	defer reconcileTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.logStep("ingest_failures", r.ingestWorker.ProcessFailures(ctx))
		case <-reconcileTicker.C:
			r.logStep("ingest_reconcile", r.ingestWorker.ProcessReconcile(ctx))
		}
	}
}

func (r *Runner) relayLoop(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.syncTick += r.interval
			r.relayTick(ctx)
			if r.syncTick >= r.syncEvery {
				r.syncTick = 0
				r.logStep("org_sync", r.processOrgSync(ctx))
			}
		}
	}
}

func (r *Runner) relayTick(ctx context.Context) {
	r.logStep("outbox_relay", r.processRelayOutbox(ctx))
	if !r.cfg.NewAPIEnabled {
		return
	}
	currentMonth := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, r.cfg.Clock()).String()
	if currentMonth != r.lastRebalanceMonth {
		r.lastRebalanceMonth = currentMonth
		r.logStep("monthly_rebalance", r.processMonthlyRebalance(ctx))
	}
	r.logStep("wallet_sync", r.processWalletSync(ctx))
	r.logStep("wallet_reconcile", r.processWalletReconcile(ctx))
	r.logStep("rebalance", r.processRebalance(ctx))
	r.logStep("overrun", r.processOverrun(ctx))
}

func (r *Runner) logStep(step string, err error) {
	if err != nil {
		r.logger.Warn("worker step failed", "step", step, "error", err)
	}
}

func (r *Runner) processMonthlyRebalance(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	return r.relaySync.EnqueueRebalanceAxis(workerCtx, "company", fmt.Sprintf("%d", r.cfg.DefaultCompanyID))
}

func (r *Runner) RunOnce(ctx context.Context) {
	r.relayTick(ctx)
	if r.cfg.IngestEnabled() {
		r.logStep("ingest_failures", r.ingestWorker.ProcessFailures(ctx))
	}
}

func (r *Runner) RunReconcileOnce(ctx context.Context) error {
	return r.ingestWorker.ProcessReconcile(ctx)
}

func (r *Runner) markRelayOutboxRetry(ctx context.Context, id string, delay time.Duration, reason string) {
	if err := r.relayJobs.MarkRelayOutboxRetry(ctx, id, delay, reason); err != nil {
		r.logger.Warn("mark relay outbox retry failed", "id", id, "error", err)
	}
}

func (r *Runner) markRelayOutboxDone(ctx context.Context, id string) {
	if err := r.relayJobs.MarkRelayOutboxDone(ctx, id); err != nil {
		r.logger.Warn("mark relay outbox done failed", "id", id, "error", err)
	}
}

func (r *Runner) markRelayOutboxFailed(ctx context.Context, id string, reason string) {
	if err := r.relayJobs.MarkRelayOutboxFailed(ctx, id, reason); err != nil {
		r.logger.Warn("mark relay outbox failed", "id", id, "error", err)
	}
}

func (r *Runner) markRebalanceDone(ctx context.Context, id string) {
	if err := r.relayJobs.MarkRebalanceDone(ctx, id); err != nil {
		r.logger.Warn("mark rebalance done failed", "id", id, "error", err)
	}
}
