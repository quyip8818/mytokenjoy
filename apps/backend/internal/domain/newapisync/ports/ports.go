package ports

import "context"

type SyncJob struct {
	CompanyID     int64
	SubKind       string
	PlatformKeyID string
	ProviderKeyID string
	DepartmentID  string
}

type SyncJobEnqueuer interface {
	InsertNewAPISync(ctx context.Context, job SyncJob) error
	InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error
}

type noopSyncJobEnqueuer struct{}

func (noopSyncJobEnqueuer) InsertNewAPISync(context.Context, SyncJob) error { return nil }
func (noopSyncJobEnqueuer) InsertRebalance(context.Context, int64, string, string) error {
	return nil
}

// NoopSyncJobEnqueuer is the default when River sync jobs are disabled.
var NoopSyncJobEnqueuer SyncJobEnqueuer = noopSyncJobEnqueuer{}

var _ SyncJobEnqueuer = noopSyncJobEnqueuer{}
