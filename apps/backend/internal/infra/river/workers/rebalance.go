package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type RebalanceWorker struct {
	river.WorkerDefaults[jobs.RebalanceArgs]
	rebalance domainbudget.Rebalancer
}

func NewRebalanceWorker(rebalance domainbudget.Rebalancer) *RebalanceWorker {
	return &RebalanceWorker{rebalance: rebalance}
}

func (w *RebalanceWorker) Work(ctx context.Context, job *river.Job[jobs.RebalanceArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	return w.rebalance.ProcessAxis(entryCtx, job.Args.AxisKind, job.Args.AxisID)
}
