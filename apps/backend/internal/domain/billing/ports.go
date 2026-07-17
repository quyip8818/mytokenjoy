package billing

import (
	"context"

	"github.com/google/uuid"
	// JobEnqueuer enqueues billing-domain River jobs without coupling to infra/jobs.
)

type JobEnqueuer interface {
	InsertWalletSync(ctx context.Context, companyID uuid.UUID) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertWalletSync(context.Context, uuid.UUID) error { return nil }

// NoopJobEnqueuer is the default when async billing jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
