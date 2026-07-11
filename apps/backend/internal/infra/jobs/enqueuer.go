package jobs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/store"
)

type Enqueuer interface {
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error
	InsertInTx(ctx context.Context, tx store.Tx, args river.JobArgs, opts *river.InsertOpts) error
}

type riverEnqueuer struct {
	client *river.Client[pgx.Tx]
}

func NewEnqueuer(client *river.Client[pgx.Tx]) Enqueuer {
	return &riverEnqueuer{client: client}
}

func (e *riverEnqueuer) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error {
	if e.client == nil {
		return fmt.Errorf("river client not configured")
	}
	_, err := e.client.Insert(ctx, args, opts)
	return err
}

func (e *riverEnqueuer) InsertInTx(ctx context.Context, tx store.Tx, args river.JobArgs, opts *river.InsertOpts) error {
	if e.client == nil {
		return fmt.Errorf("river client not configured")
	}
	_, err := e.client.InsertTx(ctx, tx.PgxTx(), args, opts)
	return err
}

var _ Enqueuer = (*riverEnqueuer)(nil)

type NoopEnqueuer struct{}

func (NoopEnqueuer) Insert(context.Context, river.JobArgs, *river.InsertOpts) error { return nil }
func (NoopEnqueuer) InsertInTx(context.Context, store.Tx, river.JobArgs, *river.InsertOpts) error {
	return nil
}

var _ Enqueuer = NoopEnqueuer{}
