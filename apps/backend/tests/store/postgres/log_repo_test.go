package postgres_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

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

func TestIngestMetricsCounts(t *testing.T) {
	t.Parallel()
	st := newIngestStore(t)
	ctx := testutil.Ctx()

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(50, 1))
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(60, 1))
	if err := st.Logs().SetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume, 50); err != nil {
		t.Fatal(err)
	}

	gaps, err := st.Logs().CountConsumeLogsAfter(ctx, 50)
	if err != nil || gaps != 1 {
		t.Fatalf("gaps = %d err=%v", gaps, err)
	}
	lag, err := st.Logs().IngestLagSeconds(ctx, 50)
	if err != nil || lag < 0 {
		t.Fatalf("lag = %d err=%v", lag, err)
	}
}
