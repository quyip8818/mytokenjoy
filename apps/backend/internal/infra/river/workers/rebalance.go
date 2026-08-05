package workers

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type RebalanceWorker struct {
	river.WorkerDefaults[jobs.RebalanceArgs]
	rebalance domainbudget.Rebalancer
	budget    domainbudget.Service
	store     store.Store
	cfg       config.Config
	clk       clock.Clock
}

func NewRebalanceWorker(rebalance domainbudget.Rebalancer, budget domainbudget.Service, st store.Store, cfg config.Config, clk clock.Clock) *RebalanceWorker {
	return &RebalanceWorker{rebalance: rebalance, budget: budget, store: st, cfg: cfg, clk: clock.OrDefault(clk)}
}

func (w *RebalanceWorker) Work(ctx context.Context, job *river.Job[jobs.RebalanceArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	if err := w.rebalance.ProcessAxis(entryCtx, job.Args.AxisKind, job.Args.AxisID); err != nil {
		return cancelIfNonRetryable(err)
	}
	if job.Args.AxisKind != store.RebalanceAxisCompany {
		return nil
	}
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, w.clk).String()
	tbs, err := w.store.TenantBackgroundState().Get(entryCtx, job.Args.CompanyID)
	if err != nil {
		return err
	}
	if tbs != nil && tbs.LastRebalancedPeriod == current {
		return nil
	}
	// Month transition: archive previous month's live budget as a snapshot.
	if err := w.budget.ArchivePreviousPeriod(entryCtx); err != nil {
		slog.WarnContext(entryCtx, "budget archive previous period failed", "err", err)
		// Non-fatal: don't block rebalance if archive fails.
	}
	return w.store.TenantBackgroundState().SetLastRebalancedPeriod(entryCtx, job.Args.CompanyID, current)
}
