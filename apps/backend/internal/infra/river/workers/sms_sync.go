package workers

import (
	"context"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

// SMSSyncExecutor abstracts the smssync.Worker.Execute method.
type SMSSyncExecutor interface {
	Execute(ctx context.Context) error
}

type SMSSyncWorker struct {
	river.WorkerDefaults[jobs.SMSSyncArgs]
	executor  SMSSyncExecutor
	companyID uuid.UUID
}

func NewSMSSyncWorker(executor SMSSyncExecutor, companyID uuid.UUID) *SMSSyncWorker {
	return &SMSSyncWorker{executor: executor, companyID: companyID}
}

func (w *SMSSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.SMSSyncArgs]) error {
	ctx = ctxcompany.With(ctx, ctxcompany.Info{CompanyID: w.companyID})
	return w.executor.Execute(ctx)
}
