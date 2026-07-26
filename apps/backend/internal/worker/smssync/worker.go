// Package smssync implements a worker that pulls model/channel/pricing data
// from the SMS system and writes it into the local NewAPI and models store.
package smssync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/integration/sms"
)

// CatalogFetcher abstracts the SMS client for testability.
type CatalogFetcher interface {
	FetchCatalog(ctx context.Context) (*sms.Catalog, error)
}

// SyncTarget abstracts the write operations (NewAPI + local DB).
type SyncTarget interface {
	UpsertChannel(ctx context.Context, ch sms.CatalogChannel) error
	UpsertModelRatio(ctx context.Context, modelID string, inputPrice, outputPrice float64) error
	UpsertModel(ctx context.Context, model sms.CatalogModel) error
	RebuildAbilities(ctx context.Context) error
}

// Worker periodically syncs data from SMS to the local system.
type Worker struct {
	fetcher  CatalogFetcher
	target   SyncTarget
	interval time.Duration
}

func New(fetcher CatalogFetcher, target SyncTarget) *Worker {
	return &Worker{fetcher: fetcher, target: target}
}

func NewWithInterval(fetcher CatalogFetcher, target SyncTarget, interval time.Duration) *Worker {
	return &Worker{fetcher: fetcher, target: target, interval: interval}
}

// Run blocks until ctx is cancelled. Syncs once immediately, then on interval.
func (w *Worker) Run(ctx context.Context) {
	w.syncOnce(ctx)
	if w.interval <= 0 {
		return
	}
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

// Execute performs a single sync cycle. Returns error if SMS is unreachable.
func (w *Worker) Execute(ctx context.Context) error {
	catalog, err := w.fetcher.FetchCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetch sms catalog: %w", err)
	}

	// Sync channels
	for _, ch := range catalog.Channels {
		if err := w.target.UpsertChannel(ctx, ch); err != nil {
			slog.Error("smssync: upsert channel failed", "channel", ch.Name, "error", err)
		}
	}
	if len(catalog.Channels) > 0 {
		if err := w.target.RebuildAbilities(ctx); err != nil {
			slog.Error("smssync: rebuild abilities failed", "error", err)
		}
	}

	// Sync pricing
	for _, m := range catalog.Models {
		if err := w.target.UpsertModelRatio(ctx, m.ModelID, m.InputPrice, m.OutputPrice); err != nil {
			slog.Error("smssync: upsert model ratio failed", "model", m.ModelID, "error", err)
		}
	}

	// Sync model metadata
	for _, m := range catalog.Models {
		if err := w.target.UpsertModel(ctx, m); err != nil {
			slog.Error("smssync: upsert model metadata failed", "model", m.ModelID, "error", err)
		}
	}

	return nil
}

func (w *Worker) syncOnce(ctx context.Context) {
	if err := w.Execute(ctx); err != nil {
		slog.Error("smssync: sync failed", "error", err)
	}
}
