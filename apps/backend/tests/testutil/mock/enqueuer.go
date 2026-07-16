package mock

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

// FailingEnqueuer is a jobs.Enqueuer that always returns an error.
// Use it to verify that enqueue failures cause transaction rollback.
type FailingEnqueuer struct{}

func (FailingEnqueuer) Insert(context.Context, river.JobArgs, *river.InsertOpts) error {
	return fmt.Errorf("enqueue failed")
}

func (FailingEnqueuer) InsertInTx(context.Context, store.Tx, river.JobArgs, *river.InsertOpts) error {
	return fmt.Errorf("enqueue failed")
}

var _ jobs.Enqueuer = FailingEnqueuer{}
