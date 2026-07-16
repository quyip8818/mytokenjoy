//go:build testhook

package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunIngestOnce starts the River client if not already running and drains all pending jobs.
// This replaces the old ingest.Worker.RunPendingOnce for tests.
func (a *App) RunIngestOnce(ctx context.Context) error {
	if a == nil || a.Workers == nil || a.Workers.river == nil {
		return nil
	}

	pool := postgresPool(a.Store)
	if pool == nil {
		return nil
	}

	// Start River if not already started (idempotent — returns nil if already running).
	_ = a.Workers.river.Start(ctx)

	// Wait for pending jobs to drain.
	deadline := time.Now().Add(5 * time.Second)
	var sawPending bool
	for time.Now().Before(deadline) {
		n := countPendingRiverJobs(ctx, pool)
		if n > 0 {
			sawPending = true
		}
		if sawPending && n == 0 {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func countPendingRiverJobs(ctx context.Context, pool *pgxpool.Pool) int {
	var count int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE state IN ('available', 'retryable', 'scheduled', 'running')
	`).Scan(&count)
	return count
}
