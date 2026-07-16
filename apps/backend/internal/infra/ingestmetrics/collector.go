package ingestmetrics

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/tokenjoy/backend/internal/store"
)

const StreamNewAPIConsume = store.ReconcileStreamNewAPIConsume

type Snapshot struct {
	WebhookAcceptedTotal int64 `json:"ingest_webhook_accepted_total"`
	ReconcileGaps        int64 `json:"ingest_reconcile_gaps"`
	LagSeconds           int64 `json:"ingest_lag_seconds"`
}

type Recorder interface {
	RecordNotifySuccess()
	Refresh(ctx context.Context, logStore store.LogStore) error
	Snapshot() Snapshot
}

type collector struct {
	mu                   sync.RWMutex
	webhookAcceptedTotal atomic.Int64
	reconcileGaps        int64
	lagSeconds           int64
}

func NewCollector() Recorder {
	return &collector{}
}

func (c *collector) RecordNotifySuccess() {
	c.webhookAcceptedTotal.Add(1)
}

func (c *collector) Refresh(ctx context.Context, logStore store.LogStore) error {
	cursor, err := logStore.GetReconcileCursor(ctx, StreamNewAPIConsume)
	if err != nil {
		return err
	}
	gaps, err := logStore.CountConsumeLogsAfter(ctx, cursor)
	if err != nil {
		return err
	}
	lag, err := logStore.IngestLagSeconds(ctx, cursor)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.reconcileGaps = gaps
	c.lagSeconds = lag
	c.mu.Unlock()
	return nil
}

func (c *collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Snapshot{
		WebhookAcceptedTotal: c.webhookAcceptedTotal.Load(),
		ReconcileGaps:        c.reconcileGaps,
		LagSeconds:           c.lagSeconds,
	}
}

type noopCollector struct{}

func (noopCollector) RecordNotifySuccess() {}

func (noopCollector) Refresh(context.Context, store.LogStore) error {
	return nil
}

func (noopCollector) Snapshot() Snapshot {
	return Snapshot{}
}

func NoopCollector() Recorder {
	return noopCollector{}
}
