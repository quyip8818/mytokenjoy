package usage_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newTestQueue(t *testing.T) (domainusage.Queue, store.Store) {
	t.Helper()
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	return domainusage.NewQueue(st.Logs()), st
}

func TestQueueEnqueue(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9104)

	if err := q.Enqueue(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Status != store.IngestJobStatusPending || f.Error != "" {
		t.Fatalf("unexpected job %+v", f)
	}
}

func TestQueueEnqueueRevivesDead(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9105)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: store.IngestJobID(logID), LogID: logID, Source: types.SourceWebhook, Error: "dead",
		Status: store.IngestJobStatusDead, Attempts: 5,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Status != store.IngestJobStatusPending || f.Attempts != 0 {
		t.Fatalf("unexpected job %+v", f)
	}
}

func TestQueueRecordFailure(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9100)

	if err := q.RecordFailure(ctx, logID, types.SourceWebhook, domain.NotFound("mapping")); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Error != "mapping" {
		t.Fatalf("error = %q", f.Error)
	}
}

func TestQueueApplyRetryDone(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9101)
	id := store.IngestJobID(logID)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: id, LogID: logID, Source: types.SourceWebhook, Error: "old",
		Status: store.IngestJobStatusPending, NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.ApplyRetry(ctx, store.IngestJob{ID: id, LogID: logID}, nil); err != nil {
		t.Fatal(err)
	}
	if n := testutil.PendingIngestJobCount(t, st); n != 0 {
		t.Fatalf("expected no pending jobs, got %d", n)
	}
}

func TestQueueApplyRetryDead(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9102)
	id := store.IngestJobID(logID)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: id, LogID: logID, Source: types.SourceWebhook, Error: "old",
		Status: store.IngestJobStatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.ApplyRetry(ctx, store.IngestJob{ID: id, LogID: logID, Attempts: 0}, domain.BadRequest("permanent")); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Status != store.IngestJobStatusDead {
		t.Fatalf("status = %q", f.Status)
	}
}

func TestQueueApplyRetryBackoff(t *testing.T) {
	t.Parallel()
	q, st := newTestQueue(t)
	ctx := testutil.Ctx()
	const logID = int64(9103)
	id := store.IngestJobID(logID)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: id, LogID: logID, Source: types.SourceWebhook, Error: "old",
		Status: store.IngestJobStatusPending, Attempts: 1,
		NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.ApplyRetry(ctx, store.IngestJob{ID: id, LogID: logID, Attempts: 1}, domain.NotFound("mapping")); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Attempts != 2 || f.Status != store.IngestJobStatusPending {
		t.Fatalf("unexpected job %+v", f)
	}
}
