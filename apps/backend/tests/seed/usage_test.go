package seed_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/runtime"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestApplyUsageBucketsSeedsPostgres(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	ctx := testutil.Ctx()

	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	if count := testutil.UsageBucketCount(st); count != 6 {
		t.Fatalf("expected 6 seeded buckets, got %d", count)
	}
	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	if count := testutil.UsageBucketCount(st); count != 6 {
		t.Fatalf("expected idempotent seed to keep 6 buckets, got %d", count)
	}
}

func TestApplyUsageBucketsProducesNonZeroDashboardSummary(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	ctx := testutil.Ctx()
	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	totals, err := st.Usage().QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if totals.CostCNY <= 0 || totals.CallCount <= 0 {
		t.Fatalf("expected non-zero summary, got %+v", totals)
	}
	const expectedRootConsumed = 67500.0
	const tolerance = 1.0
	if totals.CostCNY < expectedRootConsumed-tolerance || totals.CostCNY > expectedRootConsumed+tolerance {
		t.Fatalf("expected cost near %.0f, got %.2f", expectedRootConsumed, totals.CostCNY)
	}
}
