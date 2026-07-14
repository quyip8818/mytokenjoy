package seed_test

import (
	"math"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
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
	count := testutil.UsageBucketCount(st)
	if count == 0 {
		t.Fatal("expected seeded usage buckets")
	}
	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatal(err)
	}
	if got := testutil.UsageBucketCount(st); got != count {
		t.Fatalf("expected idempotent seed to keep %d buckets, got %d", count, got)
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
	if totals.Cost <= 0 || totals.CallCount <= 0 {
		t.Fatalf("expected non-zero summary, got %+v", totals)
	}
	if totals.DisplayCost <= 0 {
		t.Fatalf("expected non-zero display spend, got %+v", totals)
	}
	wantDisplay := totals.Cost / float64(common.DefaultPointsPerUnit)
	if math.Abs(totals.DisplayCost-wantDisplay) > 0.01 {
		t.Fatalf("display_cost should be point/PPU: cost=%v display=%v want≈%v", totals.Cost, totals.DisplayCost, wantDisplay)
	}
}
