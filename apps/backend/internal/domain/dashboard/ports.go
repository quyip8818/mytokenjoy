package dashboard

import (
	"context"

	"github.com/google/uuid"
)

type JobEnqueuer interface {
	InsertDashboardProject(ctx context.Context, companyID uuid.UUID) error
	InsertDashboardReconcile(ctx context.Context, companyID uuid.UUID) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertDashboardProject(context.Context, uuid.UUID) error   { return nil }
func (noopJobEnqueuer) InsertDashboardReconcile(context.Context, uuid.UUID) error { return nil }

// NoopJobEnqueuer is the default when async dashboard jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
