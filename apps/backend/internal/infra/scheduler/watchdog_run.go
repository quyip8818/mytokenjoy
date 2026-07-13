package scheduler

import (
	"context"
	"errors"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

var ErrDrainTimeout = errors.New("scheduler: drain river jobs timeout")

// RunOnce collects L2 due work and enqueues per-tenant jobs (same logic as tenant_watchdog Worker).
func RunOnce(ctx context.Context, cfg config.Config, st store.Store, enqueuer jobs.Enqueuer) error {
	svc := NewService(cfg, st)
	bulk := NewBulkEnqueuer(cfg, enqueuer)
	now := clock.NowUTC(cfg.Clock())
	due, err := svc.CollectDue(ctx, now)
	if err != nil {
		return err
	}
	return bulk.EnqueueDue(ctx, st, due, now)
}

// DrainActiveJobs blocks until no active river_job rows remain or timeout elapses.
func DrainActiveJobs(ctx context.Context, st store.Store, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var sawPending bool
	for time.Now().Before(deadline) {
		n, err := st.RiverJob().CountRunnableJobs(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			sawPending = true
		}
		if sawPending && n == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	n, err := st.RiverJob().CountRunnableJobs(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	return ErrDrainTimeout
}
