package budget

import "context"

// JobEnqueuer enqueues budget-domain River jobs without coupling to infra/jobs.
type JobEnqueuer interface {
	InsertBudgetProjection(ctx context.Context, companyID int64) error
	InsertOverrun(ctx context.Context, companyID int64, payload []byte) error
	InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error
	InsertBudgetReconcile(ctx context.Context, companyID int64) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertBudgetProjection(context.Context, int64) error { return nil }
func (noopJobEnqueuer) InsertOverrun(context.Context, int64, []byte) error {
	return nil
}
func (noopJobEnqueuer) InsertRebalance(context.Context, int64, string, string) error { return nil }
func (noopJobEnqueuer) InsertBudgetReconcile(context.Context, int64) error           { return nil }

// NoopJobEnqueuer is the default when async budget jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
