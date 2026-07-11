package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type MonthlyRebalanceWorker struct {
	river.WorkerDefaults[jobs.MonthlyRebalanceArgs]
	scheduler *domainbudget.MonthlyRebalanceScheduler
}

func NewMonthlyRebalanceWorker(scheduler *domainbudget.MonthlyRebalanceScheduler) *MonthlyRebalanceWorker {
	return &MonthlyRebalanceWorker{scheduler: scheduler}
}

func (w *MonthlyRebalanceWorker) Work(ctx context.Context, _ *river.Job[jobs.MonthlyRebalanceArgs]) error {
	if w.scheduler == nil {
		return nil
	}
	return w.scheduler.EnqueueMonthlyRebalanceAll(ctx)
}
