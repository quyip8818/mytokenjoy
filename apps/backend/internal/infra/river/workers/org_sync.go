package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type OrgSyncWorker struct {
	river.WorkerDefaults[jobs.OrgSyncArgs]
	sync domainorg.SyncService
}

func NewOrgSyncWorker(sync domainorg.SyncService) *OrgSyncWorker {
	return &OrgSyncWorker{sync: sync}
}

func (w *OrgSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.OrgSyncArgs]) error {
	if w.sync == nil {
		return nil
	}
	return w.sync.RunScheduledSyncAll(ctx)
}
