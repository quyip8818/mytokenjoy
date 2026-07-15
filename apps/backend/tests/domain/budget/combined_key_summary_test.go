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

func TestBudgetProjectorWritesCombinedKeySummary(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4201, 99))
	if err := ingest.IngestByLogID(ctx, 4201, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projector := budget.NewAsync(cfg, st, riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st), budget.NoopCombinedKeyCache, logger).Projector
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	remain, version := budgetfix.CombinedKeyRemain(t, st, contract.IDPlatformKey1)
	if remain == nil || *remain <= 0 {
		t.Fatalf("expected positive gateway soft remain, got %v", remain)
	}
	if version != 1 {
		t.Fatalf("expected version 1, got %d", version)
	}
}

func TestBudgetProjectorVersionMonotonicPerBatch(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4211, 99))
	if err := ingest.IngestByLogID(ctx, 4211, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projector := budget.NewAsync(cfg, st, riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st), budget.NoopCombinedKeyCache, logger).Projector
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if got := budgetfix.CombinedKeyRemainVersion(t, st, contract.IDPlatformKey1); got != 1 {
		t.Fatalf("expected version 1 after first batch, got %d", got)
	}

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4212, 99))
	if err := ingest.IngestByLogID(ctx, 4212, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if got := budgetfix.CombinedKeyRemainVersion(t, st, contract.IDPlatformKey1); got != 2 {
		t.Fatalf("expected version 2 after second batch, got %d", got)
	}
}

func TestComputeGatewaySummaryUpdatesSkipsMissingMapping(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	updates, err := budget.ComputeGatewaySummaryUpdates(ctx, st, map[string]struct{}{
		"missing-key": {},
	}, cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected no updates, got %v", updates)
	}
}

func TestBudgetReconcileRefreshesCombinedKeySummary(t *testing.T) {
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
	budgetAsync := budget.NewAsync(cfg, st, riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st), budget.NoopCombinedKeyCache, logger)
	if _, err := budgetAsync.Projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1, 0.01)
	if err := budgetAsync.Reconcile.RunCompany(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

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
	got := budgetfix.PlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1)
	if budget.ConsumedDrift(wantPlatformKey, got) {
		t.Fatalf("expected repaired consumed %v, got %v", wantPlatformKey, got)
	}
	if version := budgetfix.CombinedKeyRemainVersion(t, st, contract.IDPlatformKey1); version < 2 {
		t.Fatalf("expected summary version to increase after reconcile, got %d", version)
	}
}
