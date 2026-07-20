package dashboard_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCostSummaryUsesDisplayCostNotPointsOrPPU(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)

	// Paid lot-priced spend: 5000 point at unit_price 0.002 => 10 display (not 5000/1000=5).
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: contract.IDDept3, MemberID: contract.IDMember1,
		Model: "gpt-4o", QuotaConsumed: 5000, DisplayCost: 10, CallCount: 2,
	}); err != nil {
		t.Fatal(err)
	}
	// Gift/overdraft segment: points consume with zero display spend.
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: contract.IDDept3, MemberID: contract.IDMember1,
		Model: "gpt-4o-mini", QuotaConsumed: 3000, DisplayCost: 0, CallCount: 1,
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := svc.CostSummary(ctx, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, uuid.Nil, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalCost != 10 {
		t.Fatalf("expected display total 10, got %v", summary.TotalCost)
	}
	ppuApprox := (5000 + 3000) / float64(common.DefaultQuotaPerUnit)
	if summary.TotalCost == ppuApprox {
		t.Fatalf("totalCost must not equal point/PPU approximation %v", ppuApprox)
	}

	series, err := svc.UsageSeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       bucket,
		End:         bucket.Add(time.Hour),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	if len(series.Points) != 1 || series.Points[0].Cost != 10 {
		t.Fatalf("expected hour series display cost 10, got %+v", series.Points)
	}
}
