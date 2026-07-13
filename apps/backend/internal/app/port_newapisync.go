package app

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type newAPISyncJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

func NewNewAPISyncEnqueuer(enqueuer jobs.Enqueuer) ports.SyncJobEnqueuer {
	return newAPISyncJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (n newAPISyncJobEnqueuer) InsertNewAPISync(ctx context.Context, job ports.SyncJob) error {
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

var _ ports.SyncJobEnqueuer = newAPISyncJobEnqueuer{}
