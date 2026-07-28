package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/worker/smssync"
)

// SMSSyncWorker is a thin River adapter that delegates to SMSSyncExecutor.
type SMSSyncWorker struct {
	river.WorkerDefaults[jobs.SMSSyncArgs]
	executor *smssync.SMSSyncExecutor
}

func NewSMSSyncWorker(executor *smssync.SMSSyncExecutor) *SMSSyncWorker {
	return &SMSSyncWorker{executor: executor}
}

func (w *SMSSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.SMSSyncArgs]) error {
	return w.executor.Execute(ctx)
}
