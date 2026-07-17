package budget

import (
	"context"

	"github.com/google/uuid"
	// JobEnqueuer enqueues budget-domain River jobs without coupling to infra/jobs.
)

type JobEnqueuer interface {
	InsertOverrun(ctx context.Context, companyID uuid.UUID, payload []byte) error
	InsertRebalance(ctx context.Context, companyID uuid.UUID, axisKind, axisID string) error
	InsertBudgetReconcile(ctx context.Context, companyID uuid.UUID) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertOverrun(context.Context, uuid.UUID, []byte) error           { return nil }
func (noopJobEnqueuer) InsertRebalance(context.Context, uuid.UUID, string, string) error { return nil }
func (noopJobEnqueuer) InsertBudgetReconcile(context.Context, uuid.UUID) error           { return nil }

// NoopJobEnqueuer is the default when async budget jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
