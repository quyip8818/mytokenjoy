package remote

import (
	"context"
	"time"
)

type JobEnqueuer interface {
	InsertOrgSync(ctx context.Context, companyID int64, scheduledAt *time.Time) error
	CancelPendingOrgSync(ctx context.Context, companyID int64) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertOrgSync(context.Context, int64, *time.Time) error { return nil }
func (noopJobEnqueuer) CancelPendingOrgSync(context.Context, int64) error      { return nil }

// NoopJobEnqueuer is the default when org sync jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
