// Package pricingsync implements a background worker that pulls pricing data
// from the official management platform and writes it into the local NewAPI instance.
package pricingsync

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/integration/platform"
)

// Worker periodically syncs pricing from the platform to the local NewAPI.
type Worker struct {
	platform    *platform.Client
	adminport   adminport.Port
	interval    time.Duration
	lastVersion string
}

func New(p *platform.Client, a adminport.Port, interval time.Duration) *Worker {
	return &Worker{platform: p, adminport: a, interval: interval}
}

// Run blocks until ctx is cancelled. Syncs once immediately, then on interval.
func (w *Worker) Run(ctx context.Context) {
	w.syncOnce(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.syncOnce(ctx)
		}
	}
}

// SyncNow triggers a single sync cycle (used by manual trigger handler).
func (w *Worker) SyncNow(ctx context.Context) error {
	return w.doSync(ctx)
}

// LastVersion returns the version string of the last successful sync.
func (w *Worker) LastVersion() string { return w.lastVersion }

func (w *Worker) syncOnce(ctx context.Context) {
	if err := w.doSync(ctx); err != nil {
		slog.Warn("pricing sync failed", "error", err)
	}
}

func (w *Worker) doSync(ctx context.Context) error {
	latest, err := w.platform.GetLatestPricing(ctx)
	if err != nil {
		return err
	}
	if latest.Version == w.lastVersion {
		return nil
	}

	// 全量替换 — 平台是 SOT
	if err := w.adminport.UpdateOption(ctx, "ModelRatio", latest.ModelRatioJSON); err != nil {
		return err
	}
	if err := w.adminport.UpdateOption(ctx, "CompletionRatio", latest.CompletionRatioJSON); err != nil {
		return err
	}

	w.lastVersion = latest.Version
	slog.Info("pricing synced", "version", latest.Version)
	return nil
}
