package ingestmetrics

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/tokenjoy/backend/internal/store"
)

const StreamNewAPIConsume = store.ReconcileStreamNewAPIConsume

type Snapshot struct {
	NotifyTotal       int64 `json:"ingest_notify_total"`
	ReconcileGaps     int64 `json:"ingest_reconcile_gaps"`
	FailuresPending   int   `json:"ingest_failures_pending"`
	LagSeconds        int64 `json:"ingest_lag_seconds"`
}

type Recorder interface {
	RecordNotifySuccess()
	Refresh(ctx context.Context, logStore store.LogStore) error
	Snapshot() Snapshot
}

type collector struct {
	mu              sync.RWMutex
	notifyTotal     atomic.Int64
	reconcileGaps   int64
	failuresPending int
	lagSeconds      int64
}

func NewCollector() Recorder {
	return &collector{}
}

func (c *collector) RecordNotifySuccess() {
	c.notifyTotal.Add(1)
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
	pending, err := logStore.CountPendingIngestFailures(ctx)
	if err != nil {
		return err
	}
	lag, err := logStore.IngestLagSeconds(ctx, cursor)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.reconcileGaps = gaps
	c.failuresPending = pending
	c.lagSeconds = lag
	c.mu.Unlock()
	return nil
}

func (c *collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Snapshot{
		NotifyTotal:     c.notifyTotal.Load(),
		ReconcileGaps:   c.reconcileGaps,
		FailuresPending: c.failuresPending,
		LagSeconds:      c.lagSeconds,
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
