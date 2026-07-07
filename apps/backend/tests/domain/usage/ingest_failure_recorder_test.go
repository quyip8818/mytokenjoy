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

func TestFailureRecorderRecordFailure(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	recorder := domainusage.NewFailureRecorder(st.Logs(), nil)
	const logID = int64(9100)

	err := recorder.RecordFailure(ctx, logID, types.SourceWebhook, domain.NotFound("mapping"))
	if err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestFailure(t, st, logID, types.SourceWebhook)
	if f.Error != "mapping" {
		t.Fatalf("error = %q", f.Error)
	}
}

func TestFailureRecorderApplyRetryDone(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	recorder := domainusage.NewFailureRecorder(st.Logs(), nil)
	const logID = int64(9101)
	id := store.IngestFailureID(logID)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: id, LogID: logID, Source: types.SourceRetry, Error: "old",
		Status: store.IngestFailureStatusPending, NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	if err := recorder.ApplyRetry(ctx, store.IngestFailure{ID: id, LogID: logID}, nil); err != nil {
		t.Fatal(err)
	}
	if n := testutil.PendingIngestFailureCount(t, st); n != 0 {
		t.Fatalf("expected no pending failures, got %d", n)
	}
}

func TestFailureRecorderApplyRetryDead(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	recorder := domainusage.NewFailureRecorder(st.Logs(), nil)
	const logID = int64(9102)
	id := store.IngestFailureID(logID)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: id, LogID: logID, Source: types.SourceRetry, Error: "old",
		Status: store.IngestFailureStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	err := recorder.ApplyRetry(ctx, store.IngestFailure{ID: id, LogID: logID, Attempts: 0}, domain.BadRequest("permanent"))
	if err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestFailure(t, st, logID, types.SourceRetry)
	if f.Status != store.IngestFailureStatusDead {
		t.Fatalf("status = %q", f.Status)
	}
}

func TestFailureRecorderApplyRetryBackoff(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	recorder := domainusage.NewFailureRecorder(st.Logs(), nil)
	const logID = int64(9103)
	id := store.IngestFailureID(logID)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: id, LogID: logID, Source: types.SourceRetry, Error: "old",
		Status: store.IngestFailureStatusPending, Attempts: 1,
		NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	err := recorder.ApplyRetry(ctx, store.IngestFailure{ID: id, LogID: logID, Attempts: 1}, domain.NotFound("mapping"))
	if err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestFailure(t, st, logID, types.SourceRetry)
	if f.Attempts != 2 {
		t.Fatalf("attempts = %d, want 2", f.Attempts)
	}
	if f.Status != store.IngestFailureStatusPending {
		t.Fatalf("status = %q", f.Status)
	}
}
