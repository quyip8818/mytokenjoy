package postgres_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func platformKey1Hash() string {
	return store.HashPlatformKey("pending:" + contract.IDPlatformKey1)
}

func TestGatewaySoftSummaryUpdateBatch(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	summaries, err := st.GatewaySoftSummaries().UpdateBatch(ctx, []store.GatewaySoftSummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, SoftRemain: 123.45},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].SoftRemain != 123.45 {
		t.Fatalf("remain = %v, want 123.45", summaries[0].SoftRemain)
	}
	if summaries[0].Version != 1 {
		t.Fatalf("version = %d, want 1", summaries[0].Version)
	}

	summaries, err = st.GatewaySoftSummaries().UpdateBatch(ctx, []store.GatewaySoftSummaryUpdate{
		{PlatformKeyID: contract.IDPlatformKey1, SoftRemain: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if summaries[0].Version != 2 {
		t.Fatalf("version = %d, want 2", summaries[0].Version)
	}
}

func TestLoadPrecheckContextReturnsSoftSummary(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	testutil.SetGatewaySoftRemain(t, st, contract.IDPlatformKey1, 42)
	row, err := st.GatewayPrecheck().LoadPrecheckContext(ctx, platformKey1Hash())
	if err != nil {
		t.Fatal(err)
	}
	if row == nil {
		t.Fatal("expected precheck row")
	}
	if row.GatewaySoftRemain == nil || *row.GatewaySoftRemain != 42 {
		t.Fatalf("soft remain = %v, want 42", row.GatewaySoftRemain)
	}
	if row.GatewaySoftVersion != 1 {
		t.Fatalf("version = %d, want 1", row.GatewaySoftVersion)
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
	if row.GatewaySoftRemain != nil {
		t.Fatalf("expected NULL soft remain, got %v", row.GatewaySoftRemain)
	}
	if row.GatewaySoftVersion != 0 {
		t.Fatalf("version = %d, want 0", row.GatewaySoftVersion)
	}
}

func TestSetPlatformKeysPreservesGatewaySoftSummary(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	testutil.SetGatewaySoftRemain(t, st, contract.IDPlatformKey1, 99)
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	remain, version := testutil.GatewaySoftRemain(t, st, contract.IDPlatformKey1)
	if remain == nil || *remain != 99 {
		t.Fatalf("soft remain = %v, want 99", remain)
	}
	if version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}
}
