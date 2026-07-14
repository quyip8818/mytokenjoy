package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type RebalanceWorker struct {
	river.WorkerDefaults[jobs.RebalanceArgs]
	rebalance domainbudget.Rebalancer
	store     store.Store
	cfg       config.Config
}

func NewRebalanceWorker(rebalance domainbudget.Rebalancer, st store.Store, cfg config.Config) *RebalanceWorker {
	return &RebalanceWorker{rebalance: rebalance, store: st, cfg: cfg}
}

func (w *RebalanceWorker) Work(ctx context.Context, job *river.Job[jobs.RebalanceArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	if err := w.rebalance.ProcessAxis(entryCtx, job.Args.AxisKind, job.Args.AxisID); err != nil {
		return cancelIfNonRetryable(err)
	}
	if job.Args.AxisKind != store.RebalanceAxisCompany {
		return nil
	}
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, w.cfg.Clock()).String()
	tbs, err := w.store.TenantBackgroundState().Get(entryCtx, job.Args.CompanyID)
	if err != nil {
		return err
	}
	if tbs != nil && tbs.LastRebalancedPeriod == current {
		return nil
	}
	return w.store.TenantBackgroundState().SetLastRebalancedPeriod(entryCtx, job.Args.CompanyID, current)
}
