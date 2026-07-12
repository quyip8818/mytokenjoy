package remote

import "context"

type JobEnqueuer interface {
	InsertOrgSync(ctx context.Context, companyID int64) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertOrgSync(context.Context, int64) error { return nil }

// NoopJobEnqueuer is the default when org sync jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
