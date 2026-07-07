package ingestmetrics_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCollectorRefreshAndSnapshot(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	collector := ingestmetrics.NewCollector()

	collector.RecordNotifySuccess()
	collector.RecordNotifySuccess()

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(100, 1))
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{
		ID: 200, TokenID: 1, Quota: 1, ModelName: "m",
		CreatedAt: time.Now().Unix() - 120,
	})
	if err := st.Logs().SetReconcileCursor(ctx, ingestmetrics.StreamNewAPIConsume, 100); err != nil {
		t.Fatal(err)
	}
	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: store.IngestFailureID(300), LogID: 300, Source: "webhook",
		Error: "pending", Status: store.IngestFailureStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	if err := collector.Refresh(ctx, st.Logs()); err != nil {
		t.Fatal(err)
	}
	snap := collector.Snapshot()
	if snap.NotifyTotal != 2 {
		t.Fatalf("notify total = %d, want 2", snap.NotifyTotal)
	}
	if snap.ReconcileGaps != 1 {
		t.Fatalf("reconcile gaps = %d, want 1", snap.ReconcileGaps)
	}
	if snap.FailuresPending != 1 {
		t.Fatalf("failures pending = %d, want 1", snap.FailuresPending)
	}
	if snap.LagSeconds < 60 {
		t.Fatalf("lag seconds = %d, want >= 60", snap.LagSeconds)
	}
}

func TestNoopCollector(t *testing.T) {
	collector := ingestmetrics.NoopCollector()
	collector.RecordNotifySuccess()
	snap := collector.Snapshot()
	if snap.NotifyTotal != 0 || snap.ReconcileGaps != 0 {
		t.Fatalf("unexpected noop snapshot %+v", snap)
	}
}
