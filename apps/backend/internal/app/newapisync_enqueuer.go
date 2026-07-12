package app

import (
	"context"

	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type newAPISyncJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

func NewNewAPISyncEnqueuer(enqueuer jobs.Enqueuer) domainnewapisync.SyncJobEnqueuer {
	return newAPISyncJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (n newAPISyncJobEnqueuer) InsertNewAPISync(ctx context.Context, job domainnewapisync.SyncJob) error {
	return jobs.InsertNewAPISync(ctx, n.enqueuer, nil, jobs.NewAPISyncArgs{
		CompanyID:     job.CompanyID,
		SubKind:       job.SubKind,
		PlatformKeyID: job.PlatformKeyID,
		ProviderKeyID: job.ProviderKeyID,
		DepartmentID:  job.DepartmentID,
	})
}

func (n newAPISyncJobEnqueuer) InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error {
	return jobs.InsertRebalance(ctx, n.enqueuer, nil, companyID, axisKind, axisID)
}

var _ domainnewapisync.SyncJobEnqueuer = newAPISyncJobEnqueuer{}
