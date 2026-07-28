package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
)

// CatalogSyncWorker is a thin River adapter that delegates to catalogsync.Executor.
type CatalogSyncWorker struct {
	river.WorkerDefaults[jobs.CatalogSyncArgs]
	executor *catalogsync.Executor
}

func NewCatalogSyncWorker(executor *catalogsync.Executor) *CatalogSyncWorker {
	return &CatalogSyncWorker{executor: executor}
}

func (w *CatalogSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.CatalogSyncArgs]) error {
	return w.executor.Execute(ctx)
}
