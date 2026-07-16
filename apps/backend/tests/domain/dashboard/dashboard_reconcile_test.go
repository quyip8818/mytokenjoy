package dashboard_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/app"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestDashboardReconcileRepairsBucketDrift(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(4401, 99))
	if err := ingest.IngestByLogID(ctx, 4401, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 1})
	if err != nil || len(entries) == 0 {
		t.Fatal("expected ledger entry")
	}
	entry := entries[0]

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dashboardEnqueuer := app.NewDashboardEnqueuer(runner.Enqueuer)
	projector := domaindashboard.NewProjector(runner.Cfg, st, dashboardEnqueuer, logger)
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	bucketStart := entry.OccurredAt.UTC().Truncate(time.Hour)
	memberID := ""
	if entry.MemberID != nil {
		memberID = *entry.MemberID
	}
	if err := st.Usage().SetBucket(ctx, types.UsageBucketRow{
		BucketStart:  bucketStart,
		DepartmentID: entry.DepartmentID,
		MemberID:     memberID,
		Model:        entry.Model,
		Cost:         0.01,
		DisplayCost:  0,
		CallCount:    1,
		InputTokens:  1,
		OutputTokens: 1,
	}); err != nil {
		t.Fatal(err)
	}

	reconcile := domaindashboard.NewReconcileService(runner.Cfg, st, dashboardEnqueuer, logger)
	if err := reconcile.RunCompany(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	buckets, err := st.Usage().ListBucketsSince(ctx, bucketStart)
	if err != nil {
		t.Fatal(err)
	}
	var repaired *types.UsageBucketRow
	for i := range buckets {
		row := buckets[i]
		if row.BucketStart.Equal(bucketStart) && row.DepartmentID == entry.DepartmentID && row.Model == entry.Model {
			repaired = &row
			break
		}
	}
	if repaired == nil {
		t.Fatalf("expected repaired bucket, got %+v", buckets)
	}
	if repaired.Cost != entry.Amount {
		t.Fatalf("expected point cost %f after reconcile, got %f", entry.Amount, repaired.Cost)
	}
	if repaired.DisplayCost != entry.DisplayAmount {
		t.Fatalf("expected display cost %f after reconcile, got %f", entry.DisplayAmount, repaired.DisplayCost)
	}
	if repaired.CallCount != 1 {
		t.Fatalf("expected call count 1 after reconcile, got %d", repaired.CallCount)
	}
}
