package workers

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/clock"
)

type WatchdogWorker struct {
	river.WorkerDefaults[jobs.TenantWatchdogArgs]
	scheduler    *scheduler.Service
	bulkEnqueuer *scheduler.BulkEnqueuer
	store        store.Store
	cfg          config.Config
	clk          clock.Clock
}

func NewWatchdogWorker(svc *scheduler.Service, bulk *scheduler.BulkEnqueuer, st store.Store, cfg config.Config, clk clock.Clock) *WatchdogWorker {
	return &WatchdogWorker{
		scheduler:    svc,
		bulkEnqueuer: bulk,
		store:        st,
		cfg:          cfg,
		clk:          clock.OrDefault(clk),
	}
}

func (w *WatchdogWorker) Work(ctx context.Context, _ *river.Job[jobs.TenantWatchdogArgs]) error {
	now := clock.NowUTC(w.clk)
	due, err := w.scheduler.CollectDue(ctx, now)
	if err != nil {
		return err
	}
	return w.bulkEnqueuer.EnqueueDue(ctx, w.store, due, now)
}
