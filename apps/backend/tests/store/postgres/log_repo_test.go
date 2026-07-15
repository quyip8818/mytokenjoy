package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestEnqueuePendingCreatesClaimableJob(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(9005)

	if err := st.Logs().EnqueuePending(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	claimed, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 || claimed[0].LogID != logID {
		t.Fatalf("claimed = %+v", claimed)
	}
	if claimed[0].Error != "" {
		t.Fatalf("expected empty error, got %q", claimed[0].Error)
	}
}

func TestEnqueuePendingRevivesDead(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(9006)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: store.IngestJobID(logID), LogID: logID, Source: types.SourceWebhook,
		Error: "dead", Status: store.IngestJobStatusDead, Attempts: 9,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Logs().EnqueuePending(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	claimed, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected revived job, got %d", len(claimed))
	}
	if claimed[0].Attempts != 0 {
		t.Fatalf("attempts = %d, want 0", claimed[0].Attempts)
	}
}

func TestIngestJobHelpers(t *testing.T) {
	t.Parallel()
	if store.IngestJobID(42) != "ij-42" {
		t.Fatalf("unexpected job id %q", store.IngestJobID(42))
	}
	f := store.IngestJobFromError(7, types.SourceReconcile, errTest("boom"))
	if f.LogID != 7 || f.Source != types.SourceReconcile || f.Error != "boom" {
		t.Fatalf("unexpected job %+v", f)
	}
}

func TestUpsertJobPreservesAttemptsOnConflict(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(881001)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID:       store.IngestJobID(logID),
		LogID:    logID,
		Source:   types.SourceWebhook,
		Error:    "first",
		Attempts: 3,
		Status:   store.IngestJobStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	if err := st.Logs().UpsertJob(ctx, store.IngestJobFromError(logID, types.SourceReconcile, errTest("second"))); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected one failure, got %d", len(claimed))
	}
	if claimed[0].Attempts != 3 {
		t.Fatalf("expected attempts preserved, got %d", claimed[0].Attempts)
	}
	if claimed[0].Error != "second" {
		t.Fatalf("expected updated error, got %q", claimed[0].Error)
	}
	if claimed[0].Source != types.SourceReconcile {
		t.Fatalf("expected updated source, got %q", claimed[0].Source)
	}
}

func TestUpsertJobDoesNotReviveDead(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(9002)
	id := store.IngestJobID(logID)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID:     id,
		LogID:  logID,
		Source: types.SourceWebhook,
		Error:  "dead",
		Status: store.IngestJobStatusDead,
	}); err != nil {
		t.Fatal(err)
	}

	if err := st.Logs().UpsertJob(ctx, store.IngestJobFromError(logID, types.SourceWebhook, errTest("retry"))); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingJobs(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected dead failure to stay unclaimed, got %d", len(claimed))
	}
}

func TestClaimPendingJobsLeasesRows(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(9003)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID:        store.IngestJobID(logID),
		LogID:     logID,
		Source:    types.SourceWebhook,
		Error:     "pending",
		Status:    store.IngestJobStatusPending,
		NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	first, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil || len(first) != 1 {
		t.Fatalf("first claim: len=%d err=%v", len(first), err)
	}
	second, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("expected leased failure to be skipped, got %d", len(second))
	}
}

func TestGetConsumeLogByIDNotFound(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	_, err := st.Logs().GetConsumeLogByID(ctx, 1)
	if !errors.Is(err, store.ErrConsumeLogNotFound) {
		t.Fatalf("expected ErrConsumeLogNotFound, got %v", err)
	}
}

func TestGetConsumeLogsByIDs(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(40, 1))
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(41, 2))
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{ID: 42, TokenID: 0, Quota: 1, ModelName: "m", CreatedAt: 1})

	logs, err := st.Logs().GetConsumeLogsByIDs(ctx, []int64{40, 41, 42, 99})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs (skip token_id=0 and missing), got %d", len(logs))
	}
	got := map[int64]bool{}
	for _, raw := range logs {
		got[raw.ID] = true
	}
	if !got[40] || !got[41] {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestListConsumeLogIDsAfterFiltersAndOrders(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{ID: 10, TokenID: 0, Quota: 1, ModelName: "m", CreatedAt: 1})
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{ID: 30, TokenID: 1, Quota: 1, ModelName: "m", CreatedAt: 1})
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{ID: 20, TokenID: 2, Quota: 1, ModelName: "m", CreatedAt: 1})

	ids, err := st.Logs().ListConsumeLogIDsAfter(ctx, 15, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != 20 || ids[1] != 30 {
		t.Fatalf("unexpected ids %v", ids)
	}
}

func TestReconcileCursorRoundTrip(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	if err := st.Logs().SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, 42); err != nil {
		t.Fatal(err)
	}
	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil || cursor != 42 {
		t.Fatalf("cursor = %d err=%v", cursor, err)
	}
}

func TestMarkJobRetryAndDone(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()
	logID := int64(9004)
	id := store.IngestJobID(logID)

	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: id, LogID: logID, Source: types.SourceWebhook, Error: "pending",
		Status: store.IngestJobStatusPending, NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingJobs(ctx, 1)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim: len=%d err=%v", len(claimed), err)
	}
	if err := st.Logs().MarkJobRetry(ctx, id, time.Minute, "retrying"); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestJob(t, st, logID, types.SourceWebhook)
	if f.Attempts != 1 || f.Error != "retrying" {
		t.Fatalf("unexpected job after retry: %+v", f)
	}
	if err := st.Logs().MarkJobDone(ctx, id); err != nil {
		t.Fatal(err)
	}
	if n := testutil.PendingIngestJobCount(t, st); n != 0 {
		t.Fatalf("expected no pending jobs, got %d", n)
	}
}

func TestIngestMetricsCounts(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(50, 1))
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(60, 1))
	if err := st.Logs().SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, 50); err != nil {
		t.Fatal(err)
	}
	if err := st.Logs().UpsertJob(ctx, store.IngestJob{
		ID: store.IngestJobID(70), LogID: 70, Source: types.SourceWebhook,
		Error: "x", Status: store.IngestJobStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	gaps, err := st.Logs().CountConsumeLogsAfter(ctx, 50)
	if err != nil || gaps != 1 {
		t.Fatalf("gaps = %d err=%v", gaps, err)
	}
	pending, err := st.Logs().CountPendingIngestJobs(ctx)
	if err != nil || pending != 1 {
		t.Fatalf("pending = %d err=%v", pending, err)
	}
	lag, err := st.Logs().IngestLagSeconds(ctx, 50)
	if err != nil || lag < 0 {
		t.Fatalf("lag = %d err=%v", lag, err)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }
