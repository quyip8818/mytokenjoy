package remote

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobEnqueuer interface {
	InsertOrgSync(ctx context.Context, companyID uuid.UUID, scheduledAt *time.Time) error
	CancelPendingOrgSync(ctx context.Context, companyID uuid.UUID) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertOrgSync(context.Context, uuid.UUID, *time.Time) error { return nil }
func (noopJobEnqueuer) CancelPendingOrgSync(context.Context, uuid.UUID) error      { return nil }

// NoopJobEnqueuer is the default when org sync jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
