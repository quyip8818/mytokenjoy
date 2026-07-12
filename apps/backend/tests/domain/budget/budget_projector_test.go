package budget_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestBudgetProjectorDedupesOverrunByPlatformKey(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	for _, logID := range []int64{4201, 4202, 4203} {
		testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(logID, 99))
		if err := ingest.IngestByLogID(ctx, logID, types.SourceWebhook); err != nil {
			t.Fatal(err)
		}
	}

	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 3 {
		t.Fatalf("expected at least 3 ledger entries, got %d", len(entries))
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	baseEnqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)
	recorder := &recordingEnqueuer{inner: baseEnqueuer}
	projector := budget.NewAsync(cfg, st, recorder, budgetcheck.Noop{}, logger).Projector
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	if got := testutil.PlatformKeySnapshotUsed(t, st, contract.IDPlatformKey1); got <= 0 {
		t.Fatalf("expected projector to apply ledger batch, consumed=%v", got)
	}
	if got := recorder.overruns; got != 1 {
		t.Fatalf("expected 1 overrun enqueue for repeated platform key, got %d", got)
	}
}
