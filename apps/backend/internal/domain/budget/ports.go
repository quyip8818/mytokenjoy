package budget

import (
	"context"

	"github.com/google/uuid"
	// JobEnqueuer enqueues budget-domain River jobs without coupling to infra/jobs.
)

type JobEnqueuer interface {
	InsertOverrun(ctx context.Context, companyID uuid.UUID, payload []byte) error
	InsertRebalance(ctx context.Context, companyID uuid.UUID, axisKind string, axisID uuid.UUID) error
	InsertBudgetReconcile(ctx context.Context, companyID uuid.UUID) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertOverrun(context.Context, uuid.UUID, []byte) error { return nil }
func (noopJobEnqueuer) InsertRebalance(context.Context, uuid.UUID, string, uuid.UUID) error {
	return nil
}
func (noopJobEnqueuer) InsertBudgetReconcile(context.Context, uuid.UUID) error { return nil }

// NoopJobEnqueuer is the default when async budget jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}

// OverrunKeyControl is consumed by the budget domain's overrun processor
// to disable keys that exceed their budget. Implemented by newapisync.
type OverrunKeyControl interface {
	Enabled() bool
	DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
}
