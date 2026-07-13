package workers

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type NewAPISyncWorker struct {
	river.WorkerDefaults[jobs.NewAPISyncArgs]
	handler newapisync.OutboxHandler
}

func NewNewAPISyncWorker(handler newapisync.OutboxHandler) *NewAPISyncWorker {
	return &NewAPISyncWorker{handler: handler}
}

func (w *NewAPISyncWorker) Work(ctx context.Context, job *river.Job[jobs.NewAPISyncArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	var err error
	switch job.Args.SubKind {
	case outbox.KindCreateKey:
		_, err = w.handler.TrySyncCreate(entryCtx, job.Args.PlatformKeyID)
	case outbox.KindUpsertChannel:
		err = w.handler.SyncUpsertProviderKey(entryCtx, job.Args.ProviderKeyID)
	case outbox.KindUpdateModelLimits:
		err = w.handler.SyncModelLimitsForDepartment(entryCtx, job.Args.DepartmentID)
	default:
		return river.JobCancel(fmt.Errorf("unknown newapi sync sub kind: %s", job.Args.SubKind))
	}
	if err != nil && outbox.IsPermanentOutboxError(err) {
		return river.JobCancel(err)
	}
	return err
}
