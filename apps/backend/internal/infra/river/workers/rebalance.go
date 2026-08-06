package workers

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type RebalanceWorker struct {
	river.WorkerDefaults[jobs.RebalanceArgs]
	rebalance domainbudget.Rebalancer
	budget    domainbudget.Service
}

func NewRebalanceWorker(rebalance domainbudget.Rebalancer, budget domainbudget.Service) *RebalanceWorker {
	return &RebalanceWorker{rebalance: rebalance, budget: budget}
}

func (w *RebalanceWorker) Work(ctx context.Context, job *river.Job[jobs.RebalanceArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)

	// Month transition: rotate (archive + apply snapshot + mark) BEFORE ProcessAxis
	// so that ProcessAxis reads the latest budget values.
	if job.Args.AxisKind == store.RebalanceAxisCompany {
		if err := w.budget.RotatePeriod(entryCtx); err != nil {
			slog.WarnContext(entryCtx, "budget rotate period failed", "err", err)
			// Non-fatal: proceed with rebalance using current budget values.
		}
	}

	if err := w.rebalance.ProcessAxis(entryCtx, job.Args.AxisKind, job.Args.AxisID); err != nil {
		return cancelIfNonRetryable(err)
	}
	return nil
}
