package dashboard_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	workerfix "github.com/tokenjoy/backend/tests/testutil/worker"
)

func TestDashboardProjectorUpsertBucketFromLedger(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	runner, st, ingest := workerfix.NewRuntime(t, stub)
	ctx := testutil.Ctx()
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	raw := testutil.DefaultConsumeLog(7201, 99)
	testutil.SeedConsumeLog(t, st, raw)
	if err := ingest.IngestByLogID(ctx, 7201, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	projector := dashboard.NewProjector(runner.Cfg, st, nil, nil)
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 1})
	if err != nil || len(entries) == 0 {
		t.Fatal("expected ledger entry")
	}
	occurred := entries[0].OccurredAt
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       occurred.Truncate(time.Hour),
		End:         occurred.Truncate(time.Hour).Add(time.Hour),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("expected one bucket, got %+v", points)
	}
	if points[0].Cost != entries[0].DisplayAmount {
		t.Fatalf("series cost should be display_amount %f, got %f", entries[0].DisplayAmount, points[0].Cost)
	}
	buckets, err := st.Usage().ListBucketsSince(ctx, occurred.Truncate(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets) == 0 {
		t.Fatal("expected usage bucket row")
	}
	found := false
	for _, b := range buckets {
		if b.Model == entries[0].Model && b.Cost == entries[0].Amount && b.DisplayCost == entries[0].DisplayAmount {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected bucket with cost=%f display_cost=%f, got %+v", entries[0].Amount, entries[0].DisplayAmount, buckets)
	}
}
