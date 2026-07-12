package budget

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

// JobEnqueuer enqueues budget-domain River jobs without coupling to infra/jobs.
type JobEnqueuer interface {
	InsertBudgetProject(ctx context.Context, companyID int64) error
	InsertOverrun(ctx context.Context, companyID int64, payload []byte) error
	InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error
	InsertBudgetReconcile(ctx context.Context, companyID int64) error
}

// Notifier sends domain notifications without coupling to infra/notification.
type Notifier interface {
	Send(ctx context.Context, notification types.Notification) error
}
