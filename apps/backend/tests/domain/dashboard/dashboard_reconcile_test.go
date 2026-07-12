package dashboard_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestDashboardReconcileRepairsBucketDrift(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := workerfix.NewRuntime(t, stub)
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
	projector := domaindashboard.NewProjector(runner.Cfg, st, runner.Enqueuer, logger)
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
		CallCount:    1,
		InputTokens:  1,
		OutputTokens: 1,
	}); err != nil {
		t.Fatal(err)
	}

	reconcile := domaindashboard.NewReconcileService(runner.Cfg, st, runner.Enqueuer, logger)
	if err := reconcile.RunCompany(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       bucketStart,
		End:         bucketStart.Add(time.Hour),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("expected one bucket after reconcile, got %+v", points)
	}
	if points[0].Cost <= 0.01 {
		t.Fatalf("expected reconcile to repair bucket cost above drift seed, got %f", points[0].Cost)
	}
	if points[0].CallCount != 1 {
		t.Fatalf("expected call count 1 after reconcile, got %d", points[0].CallCount)
	}
}
