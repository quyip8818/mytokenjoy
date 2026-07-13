package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/domain/company"
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

func (w *OrgSyncWorker) Work(ctx context.Context, job *river.Job[jobs.OrgSyncArgs]) error {
	if job.Args.CompanyID == 0 {
		return nil
	}
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	return w.sync.RunScheduledSync(entryCtx)
}
