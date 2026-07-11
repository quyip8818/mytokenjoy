package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type Runner struct {
	cfg           config.Config
	asyncJobs     store.AsyncJobsRepository
	schedulerLock store.SchedulerLockRepository
	companies     store.CompanyRepository
	newAPISync    newapisync.OutboxHandler
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
	asyncJobsRepo store.AsyncJobsRepository,
	schedulerLock store.SchedulerLockRepository,
	companies store.CompanyRepository,
	logStore store.LogStore,
	metrics ingestmetrics.Recorder,
	newAPISync newapisync.OutboxHandler,
	ingest domainusage.Ingestor,
	ingestQueue domainusage.Queue,
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
	if ingestQueue == nil {
		ingestQueue = domainusage.NewQueue(logStore)
	}
	holderID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	return &Runner{
		cfg:           cfg,
		asyncJobs:     asyncJobsRepo,
		schedulerLock: schedulerLock,
		companies:     companies,
		newAPISync:    newAPISync,
		overrun:       overrun,
		rebalance:     rebalance,
		walletSync:    billingWalletSync{svc: billingSvc},
		syncSvc:       syncSvc,
		ingestWorker: NewIngestWorker(
			cfg, logStore, ingest, ingestQueue, metrics, schedulerLock, logger, holderID, cfg.IngestReconcileInterval(),
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
	go r.asyncLoop(ctx)
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
			r.logStep("ingest_pending", r.ingestWorker.ProcessPending(ctx))
		case <-reconcileTicker.C:
			r.logStep("ingest_reconcile", r.ingestWorker.ProcessReconcile(ctx))
		}
	}
}

func (r *Runner) asyncLoop(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.syncTick += r.interval
			r.asyncTick(ctx)
			if r.syncTick >= r.syncEvery {
				r.syncTick = 0
				r.logStep("org_sync", r.processOrgSync(ctx))
			}
		}
	}
}

func (r *Runner) asyncTick(ctx context.Context) {
	r.logStep("outbox_newapi_sync", r.processNewAPISyncOutbox(ctx))
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
	return r.forEachActiveCompany(ctx, func(entryCtx context.Context, co store.Company) error {
		return r.newAPISync.EnqueueRebalanceAxis(entryCtx, store.RebalanceAxisCompany, fmt.Sprintf("%d", co.ID))
	})
}

// RunOrgSyncOnce runs scheduled org sync for every active company (for tests).
func (r *Runner) RunOrgSyncOnce(ctx context.Context) error {
	return r.processOrgSync(ctx)
}

func (r *Runner) RunOnce(ctx context.Context) {
	r.asyncTick(ctx)
	if r.cfg.IngestEnabled() {
		r.logStep("ingest_pending", r.ingestWorker.ProcessPending(ctx))
	}
}

func (r *Runner) RunReconcileOnce(ctx context.Context) error {
	return r.ingestWorker.ProcessReconcile(ctx)
}

func (r *Runner) markJobRetry(ctx context.Context, id string, delay time.Duration, reason string) {
	if err := r.asyncJobs.MarkJobRetry(ctx, id, delay, reason); err != nil {
		r.logger.Warn("mark job retry failed", "id", id, "error", err)
	}
}

func (r *Runner) markJobDone(ctx context.Context, id string) {
	if err := r.asyncJobs.MarkJobDone(ctx, id); err != nil {
		r.logger.Warn("mark job done failed", "id", id, "error", err)
	}
}

func (r *Runner) markJobFailed(ctx context.Context, id string, reason string) {
	if err := r.asyncJobs.MarkJobFailed(ctx, id, reason); err != nil {
		r.logger.Warn("mark job failed", "id", id, "error", err)
	}
}

func (r *Runner) markRebalanceDone(ctx context.Context, id string) {
	if err := r.asyncJobs.MarkRebalanceDone(ctx, id); err != nil {
		r.logger.Warn("mark rebalance done failed", "id", id, "error", err)
	}
}
