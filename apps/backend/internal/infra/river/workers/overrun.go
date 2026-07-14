package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type OverrunWorker struct {
	river.WorkerDefaults[jobs.OverrunArgs]
	overrun domainbudget.OverrunProcessor
}

func NewOverrunWorker(overrun domainbudget.OverrunProcessor) *OverrunWorker {
	return &OverrunWorker{overrun: overrun}
}

func (w *OverrunWorker) Work(ctx context.Context, job *river.Job[jobs.OverrunArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	return cancelIfNonRetryable(w.overrun.ProcessOverrunPayload(entryCtx, job.Args.Payload))
}
