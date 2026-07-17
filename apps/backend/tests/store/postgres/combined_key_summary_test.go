package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func platformKey1Hash() string {
	return store.HashPlatformKey("pending:" + contract.IDPlatformKey1.String())
}

func TestCombinedKeySummaryUpdateBatch(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	summaries, err := st.CombinedKeySummaries().UpdateBatch(ctx, []store.CombinedKeySummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, Remain: 123.45},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Remain != 123.45 {
		t.Fatalf("remain = %v, want 123.45", summaries[0].Remain)
	}
	if summaries[0].Version != 1 {
		t.Fatalf("version = %d, want 1", summaries[0].Version)
	}

	summaries, err = st.CombinedKeySummaries().UpdateBatch(ctx, []store.CombinedKeySummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, Remain: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if summaries[0].Version != 2 {
		t.Fatalf("version = %d, want 2", summaries[0].Version)
	}
}

func TestCombinedKeySummaryDecrementBatch(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	if _, err := st.CombinedKeySummaries().UpdateBatch(ctx, []store.CombinedKeySummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, Remain: 100},
	}); err != nil {
		t.Fatal(err)
	}

	summaries, err := st.CombinedKeySummaries().DecrementBatch(ctx, map[uuid.UUID]float64{
		contract.IDPlatformKey1:                                12.5,
		uuid.MustParse("00000000-0000-7000-0000-ffffffffffff"): 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 updated summary, got %d", len(summaries))
	}
	if summaries[0].Remain != 87.5 {
		t.Fatalf("remain = %v, want 87.5", summaries[0].Remain)
	}
	if summaries[0].Version != 2 {
		t.Fatalf("version = %d, want 2", summaries[0].Version)
	}
}

func TestCombinedKeySummaryDecrementBatchFloorsAtZero(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	if _, err := st.CombinedKeySummaries().UpdateBatch(ctx, []store.CombinedKeySummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, Remain: 5},
	}); err != nil {
		t.Fatal(err)
	}

	summaries, err := st.CombinedKeySummaries().DecrementBatch(ctx, map[uuid.UUID]float64{
		contract.IDPlatformKey1: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 updated summary, got %d", len(summaries))
	}
	if summaries[0].Remain != 0 {
		t.Fatalf("remain = %v, want 0", summaries[0].Remain)
	}
}

func TestCombinedKeySummaryDecrementBatchSkipsNullRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	summaries, err := st.CombinedKeySummaries().DecrementBatch(ctx, map[uuid.UUID]float64{
		contract.IDPlatformKey1: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected no updates for NULL soft remain, got %d", len(summaries))
	}
	remain, _ := budgetfix.CombinedKeyRemain(t, st, contract.IDPlatformKey1)
	if remain != nil {
		t.Fatalf("expected soft remain to stay NULL, got %v", remain)
	}
}

func TestLoadPrecheckContextReturnsSoftSummary(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	budgetfix.SetCombinedKeyRemain(t, st, contract.IDPlatformKey1, 42)
	row, err := st.GatewayPrecheck().LoadPrecheckContext(ctx, platformKey1Hash())
	if err != nil {
		t.Fatal(err)
	}
	if row == nil {
		t.Fatal("expected precheck row")
	}
	if row.CombinedKeyRemain == nil || *row.CombinedKeyRemain != 42 {
		t.Fatalf("soft remain = %v, want 42", row.CombinedKeyRemain)
	}
	if row.CombinedKeyRemainVersion != 1 {
		t.Fatalf("version = %d, want 1", row.CombinedKeyRemainVersion)
	}
}

func TestLoadPrecheckContextNullSoftSummary(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	row, err := st.GatewayPrecheck().LoadPrecheckContext(ctx, platformKey1Hash())
	if err != nil {
		t.Fatal(err)
	}
	if row == nil {
		t.Fatal("expected precheck row")
	}
	if row.CombinedKeyRemain != nil {
		t.Fatalf("expected NULL soft remain, got %v", row.CombinedKeyRemain)
	}
	if row.CombinedKeyRemainVersion != 0 {
		t.Fatalf("version = %d, want 0", row.CombinedKeyRemainVersion)
	}
}

func TestSetPlatformKeysPreservesCombinedKeySummary(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	budgetfix.SetCombinedKeyRemain(t, st, contract.IDPlatformKey1, 99)
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	remain, version := budgetfix.CombinedKeyRemain(t, st, contract.IDPlatformKey1)
	if remain == nil || *remain != 99 {
		t.Fatalf("soft remain = %v, want 99", remain)
	}
	if version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}
}
