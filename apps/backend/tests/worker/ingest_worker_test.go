package worker_test

import (
	"testing"
	"time"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestReconcileMultipleLogs(t *testing.T) {
	runner, st, _ := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()
	tokenID := int64(88)
	relayfix.UpsertMapping(t, st, relayfix.MappingOpts{
		PlatformKeyID: seed.IDPlatformKey1, NewAPITokenID: tokenID,
	})
	for _, id := range []int64{801, 802, 803} {
		testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(id, tokenID))
	}

	if err := runner.RunReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{801, 802, 803} {
		ingested, err := testutil.HasLedgerLogID(st, id)
		if err != nil || !ingested {
			t.Fatalf("expected log %d ingested, err=%v", id, err)
		}
	}
	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil || cursor != 803 {
		t.Fatalf("cursor = %d err=%v", cursor, err)
	}
}

func TestReconcileBusinessFailAdvancesCursor(t *testing.T) {
	runner, st, _ := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()
	const logID = int64(701)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(logID, 33))

	if err := runner.RunReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}

	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil || cursor != logID {
		t.Fatalf("cursor = %d err=%v", cursor, err)
	}
	testutil.AssertIngestFailure(t, st, logID, types.SourceReconcile)
	ingested, _ := testutil.HasLedgerLogID(st, logID)
	if ingested {
		t.Fatal("expected no ledger entry for business failure")
	}
}

func TestReconcileSkipsZeroTokenLogs(t *testing.T) {
	runner, st, _ := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()
	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{ID: 901, TokenID: 0, Quota: 1, ModelName: "m", CreatedAt: 1})
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(902, 44))
	relayfix.UpsertMapping(t, st, relayfix.MappingOpts{
		PlatformKeyID: seed.IDPlatformKey1, NewAPITokenID: 44,
	})

	if err := runner.RunReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}
	cursor, err := st.Logs().GetReconcileCursor(ctx, store.ReconcileStreamNewAPIConsume)
	if err != nil || cursor != 902 {
		t.Fatalf("cursor = %d err=%v", cursor, err)
	}
	ingested, _ := testutil.HasLedgerLogID(st, 901)
	if ingested {
		t.Fatal("expected zero-token log to be skipped")
	}
}

func TestReconcileRunsWithoutNewAPIEnabled(t *testing.T) {
	runner, st, _ := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()
	tokenID := int64(66)
	relayfix.UpsertMapping(t, st, relayfix.MappingOpts{
		PlatformKeyID: seed.IDPlatformKey1, NewAPITokenID: tokenID,
	})
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(750, tokenID))

	if err := runner.RunReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}
	ingested, err := testutil.HasLedgerLogID(st, 750)
	if err != nil || !ingested {
		t.Fatalf("expected reconcile ingest without relay enabled, err=%v", err)
	}
}

func TestIngestFailureMaxAttemptsMarksDead(t *testing.T) {
	runner, st, _ := workerfix.NewIngestOnlyRunner(t)
	ctx := testutil.Ctx()
	const logID = int64(602)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(logID, 88))

	if err := st.Logs().UpsertFailure(ctx, store.IngestFailure{
		ID:        store.IngestFailureID(logID),
		LogID:     logID,
		Source:    types.SourceWebhook,
		Error:     "mapping missing",
		Status:    store.IngestFailureStatusPending,
		Attempts:  store.IngestFailureMaxAttempts - 1,
		NextRetry: time.Now().Add(-time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	runner.RunOnce(ctx)

	f := testutil.AssertIngestFailure(t, st, logID, "")
	if f.Status != store.IngestFailureStatusDead {
		t.Fatalf("expected dead status, got %q", f.Status)
	}
	if testutil.PendingIngestFailureCount(t, st) != 0 {
		t.Fatal("expected no pending failures after dead")
	}
}
