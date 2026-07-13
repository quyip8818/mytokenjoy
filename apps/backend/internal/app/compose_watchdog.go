package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	"github.com/tokenjoy/backend/internal/store"
)

// startDeferredWatchdog enqueues L2 due work after a short delay so HTTP /healthz is not blocked.
func startDeferredWatchdog(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	st store.Store,
	holder *jobs.Holder,
) {
	if !cfg.RiverEnabled {
		return
	}
	delay := cfg.WatchdogStartupDelay()
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		if err := scheduler.RunOnce(ctx, cfg, st, holder); err != nil {
			if ctx.Err() == nil {
				logger.Warn("deferred watchdog run failed", "error", err)
			}
			return
		}
		if ctx.Err() == nil {
			logger.Info("deferred watchdog enqueued due work", "delay", delay)
		}
	}()
}
