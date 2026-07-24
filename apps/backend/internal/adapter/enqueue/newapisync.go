package enqueue

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapisync/ports"
)

type newAPISyncJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewNewAPISyncEnqueuer adapts infra/jobs.Enqueuer to ports.SyncJobEnqueuer.
func NewNewAPISyncEnqueuer(enqueuer jobs.Enqueuer) ports.SyncJobEnqueuer {
	return newAPISyncJobEnqueuer{enqueuer: JobsOrNoop(enqueuer)}
}

func (n newAPISyncJobEnqueuer) InsertNewAPISync(ctx context.Context, job ports.SyncJob) error {
	return jobs.InsertNewAPISync(ctx, n.enqueuer, nil, jobs.NewAPISyncArgs{
		CompanyID:     job.CompanyID,
		SubKind:       job.SubKind,
		PlatformKeyID: job.PlatformKeyID,
		ProviderKeyID: job.ProviderKeyID,
	})
}

func (n newAPISyncJobEnqueuer) InsertRebalance(ctx context.Context, companyID uuid.UUID, axisKind string, axisID uuid.UUID) error {
	return jobs.InsertRebalance(ctx, n.enqueuer, nil, companyID, axisKind, axisID)
}

var _ ports.SyncJobEnqueuer = newAPISyncJobEnqueuer{}
