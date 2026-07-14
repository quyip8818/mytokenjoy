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

func TestHealUsageBucketDisplayCostsRescalesMisCopiedPoints(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: "dept-3", MemberID: "m-1",
		Model: "gpt-4o", Cost: 21000000, DisplayCost: 21000000, CallCount: 10,
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.HealUsageBucketDisplayCosts(ctx, st); err != nil {
		t.Fatal(err)
	}
	rows, err := st.Usage().QueryFilteredBuckets(ctx, types.UsageAggregateQuery{
		Start: bucket, End: bucket.Add(time.Hour), Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %+v", rows)
	}
	want := 21000000 / float64(common.DefaultPointsPerUnit)
	if math.Abs(rows[0].DisplayCost-want) > 0.01 {
		t.Fatalf("display_cost=%v want %v", rows[0].DisplayCost, want)
	}
	if rows[0].Cost != 21000000 {
		t.Fatalf("point cost must stay unchanged, got %v", rows[0].Cost)
	}
}
