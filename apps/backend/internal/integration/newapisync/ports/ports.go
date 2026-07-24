package ports

import (
	"context"

	"github.com/google/uuid"
)

type SyncJob struct {
	CompanyID     uuid.UUID
	SubKind       string
	PlatformKeyID uuid.UUID
	ProviderKeyID uuid.UUID
}

type SyncJobEnqueuer interface {
	InsertNewAPISync(ctx context.Context, job SyncJob) error
	InsertRebalance(ctx context.Context, companyID uuid.UUID, axisKind string, axisID uuid.UUID) error
}

type noopSyncJobEnqueuer struct{}

func (noopSyncJobEnqueuer) InsertNewAPISync(context.Context, SyncJob) error { return nil }
func (noopSyncJobEnqueuer) InsertRebalance(context.Context, uuid.UUID, string, uuid.UUID) error {
	return nil
}

// NoopSyncJobEnqueuer is the default when River sync jobs are disabled.
var NoopSyncJobEnqueuer SyncJobEnqueuer = noopSyncJobEnqueuer{}

var _ SyncJobEnqueuer = noopSyncJobEnqueuer{}
