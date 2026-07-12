package budget_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestBudgetReconcileRepairsDriftAndEnqueuesRebalance(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4301, 99))
	if err := ingest.IngestByLogID(ctx, 4301, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	recorder := &recordingBudgetEnqueuer{inner: riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st)}
	budgetAsync := budget.NewAsync(cfg, st, recorder, budget.NoopGatewaySoftCache, logger)
	if _, err := budgetAsync.Projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	beforeRebalance := recorder.rebalances

	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	wantMap, err := budget.ExpectedConsumed(ctx, st.Org().Nodes(), entries, cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}
	var wantPlatformKey float64
	for key, want := range wantMap {
		if key.Kind == store.AxisKindPlatformKey && key.AxisID == contract.IDPlatformKey1 {
			wantPlatformKey = want
			break
		}
	}
	if wantPlatformKey <= 0 {
		t.Fatalf("expected positive platform key consumed from ledger, got %v", wantPlatformKey)
	}

	budgetfix.SetPlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1, 0.01)

	if err := budgetAsync.Reconcile.RunCompany(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	got := budgetfix.PlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1)
	if budget.ConsumedDrift(wantPlatformKey, got) {
		t.Fatalf("expected reconcile to repair platform key consumed to %v, got %v", wantPlatformKey, got)
	}
	if recorder.rebalances <= beforeRebalance {
		t.Fatal("expected company rebalance enqueue after budget reconcile repair")
	}
}
