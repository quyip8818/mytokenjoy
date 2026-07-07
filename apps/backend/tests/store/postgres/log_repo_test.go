package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpsertFailurePreservesAttemptsOnConflict(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	logID := int64(9001)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID:       store.IngestFailureID(logID),
		LogID:    logID,
		Source:   types.SourceWebhook,
		Error:    "first",
		Attempts: 3,
		Status:   store.IngestFailureStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailureFromError(logID, types.SourceReconcile, errTest("second"))); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingFailures(ctx, 1)
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

func TestUpsertFailureDoesNotReviveDead(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	logID := int64(9002)
	id := store.IngestFailureID(logID)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID:     id,
		LogID:  logID,
		Source: types.SourceWebhook,
		Error:  "dead",
		Status: store.IngestFailureStatusDead,
	}); err != nil {
		t.Fatal(err)
	}

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailureFromError(logID, types.SourceWebhook, errTest("retry"))); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingFailures(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected dead failure to stay unclaimed, got %d", len(claimed))
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestClaimPendingFailuresLeasesRows(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	logID := int64(9003)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID:        store.IngestFailureID(logID),
		LogID:     logID,
		Source:    types.SourceWebhook,
		Error:     "pending",
		Status:    store.IngestFailureStatusPending,
		NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	first, err := st.Logs().ClaimPendingFailures(ctx, 1)
	if err != nil || len(first) != 1 {
		t.Fatalf("first claim: len=%d err=%v", len(first), err)
	}
	second, err := st.Logs().ClaimPendingFailures(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("expected leased failure to be skipped, got %d", len(second))
	}
}

func TestGetConsumeLogByIDNotFound(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()

	_, err := st.Logs().GetConsumeLogByID(ctx, 1)
	if !errors.Is(err, store.ErrConsumeLogNotFound) {
		t.Fatalf("expected ErrConsumeLogNotFound, got %v", err)
	}
}

func TestListConsumeLogIDsAfterFiltersAndOrders(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
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
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()

	if err := st.Logs().SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, 42); err != nil {
		t.Fatal(err)
	}
	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil || cursor != 42 {
		t.Fatalf("cursor = %d err=%v", cursor, err)
	}
}

func TestMarkFailureRetryAndDone(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()
	logID := int64(9004)
	id := store.IngestFailureID(logID)

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: id, LogID: logID, Source: types.SourceWebhook, Error: "pending",
		Status: store.IngestFailureStatusPending, NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	claimed, err := st.Logs().ClaimPendingFailures(ctx, 1)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim: len=%d err=%v", len(claimed), err)
	}
	next := time.Now().Add(time.Minute)
	if err := st.Logs().MarkFailureRetry(ctx, id, next, "retrying"); err != nil {
		t.Fatal(err)
	}
	f := testutil.AssertIngestFailure(t, st, logID, types.SourceWebhook)
	if f.Attempts != 1 || f.Error != "retrying" {
		t.Fatalf("unexpected failure after retry: %+v", f)
	}
	if err := st.Logs().MarkFailureDone(ctx, id); err != nil {
		t.Fatal(err)
	}
	if n := testutil.PendingIngestFailureCount(t, st); n != 0 {
		t.Fatalf("expected no pending failures, got %d", n)
	}
}

func TestIngestMetricsCounts(t *testing.T) {
	_, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ctx := testutil.Ctx()

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(50, 1))
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(60, 1))
	if err := st.Logs().SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, 50); err != nil {
		t.Fatal(err)
	}
	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID: store.IngestFailureID(70), LogID: 70, Source: types.SourceWebhook,
		Error: "x", Status: store.IngestFailureStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	gaps, err := st.Logs().CountConsumeLogsAfter(ctx, 50)
	if err != nil || gaps != 1 {
		t.Fatalf("gaps = %d err=%v", gaps, err)
	}
	pending, err := st.Logs().CountPendingIngestFailures(ctx)
	if err != nil || pending != 1 {
		t.Fatalf("pending = %d err=%v", pending, err)
	}
	lag, err := st.Logs().IngestLagSeconds(ctx, 50)
	if err != nil || lag < 0 {
		t.Fatalf("lag = %d err=%v", lag, err)
	}
}

func TestIngestFailureHelpers(t *testing.T) {
	if store.IngestFailureID(42) != "if-42" {
		t.Fatalf("unexpected failure id %q", store.IngestFailureID(42))
	}
	f := store.IngestFailureFromError(7, types.SourceReconcile, errTest("boom"))
	if f.LogID != 7 || f.Source != types.SourceReconcile || f.Error != "boom" {
		t.Fatalf("unexpected failure %+v", f)
	}
}
