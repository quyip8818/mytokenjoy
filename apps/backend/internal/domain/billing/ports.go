package billing

import "context"

// JobEnqueuer enqueues billing-domain River jobs without coupling to infra/jobs.
type JobEnqueuer interface {
	InsertWalletSync(ctx context.Context, companyID int64) error
}

type noopJobEnqueuer struct{}

func (noopJobEnqueuer) InsertWalletSync(context.Context, int64) error { return nil }

// NoopJobEnqueuer is the default when async billing jobs are disabled.
var NoopJobEnqueuer JobEnqueuer = noopJobEnqueuer{}

var _ JobEnqueuer = noopJobEnqueuer{}
