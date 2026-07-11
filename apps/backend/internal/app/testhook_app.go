//go:build testhook

package app

import "context"

func (a *App) RunIngestOnce(ctx context.Context) error {
	if a == nil || a.Workers == nil || a.Workers.ingest == nil {
		return nil
	}
	return a.Workers.ingest.RunPendingOnce(ctx)
}
