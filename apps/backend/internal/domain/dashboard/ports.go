package dashboard

import "context"

type JobEnqueuer interface {
	InsertDashboardProject(ctx context.Context, companyID int64) error
	InsertDashboardReconcile(ctx context.Context, companyID int64) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertDashboardProject(context.Context, int64) error   { return nil }
func (noopJobEnqueuer) InsertDashboardReconcile(context.Context, int64) error { return nil }

// NoopJobEnqueuer is the default when async dashboard jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
