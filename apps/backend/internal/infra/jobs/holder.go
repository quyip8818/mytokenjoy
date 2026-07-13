package jobs

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/store"
)

// Holder allows wiring domain enqueue paths before the River client is constructed.
// Registry builds with NoopEnqueuer first; compose_worker sets the real enqueuer after NewClient.
type Holder struct {
	inner Enqueuer
}

func NewHolder(initial Enqueuer) *Holder {
	if initial == nil {
		initial = NoopEnqueuer{}
	}
	return &Holder{inner: initial}
}

func (h *Holder) Set(e Enqueuer) {
	if e == nil {
		e = NoopEnqueuer{}
	}
	h.inner = e
}

func (h *Holder) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error {
	return h.inner.Insert(ctx, args, opts)
}

func (h *Holder) InsertInTx(ctx context.Context, tx store.Tx, args river.JobArgs, opts *river.InsertOpts) error {
	return h.inner.InsertInTx(ctx, tx, args, opts)
}

var _ Enqueuer = (*Holder)(nil)
