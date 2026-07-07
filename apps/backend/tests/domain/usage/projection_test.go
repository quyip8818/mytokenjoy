package usage_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestProjectionUpsertBucketIdempotent(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	occurred := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	entry := types.UsageLedgerEntry{
		ID: "ledger-1", PlatformKeyID: "plk-1", DepartmentID: "dept-3",
		Model: "gpt-4o", AmountCNY: 1.5, OccurredAt: occurred,
		InputTokens: 100, OutputTokens: 50,
	}
	if err := usage.Apply(ctx, st, entry); err != nil {
		t.Fatal(err)
	}
	entry2 := entry
	entry2.ID = "ledger-2"
	entry2.AmountCNY = 2.0
	if err := usage.Apply(ctx, st, entry2); err != nil {
		t.Fatal(err)
	}
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
	want := 3.5
	if points[0].CostCNY != want {
		t.Fatalf("expected aggregated cost %f, got %f", want, points[0].CostCNY)
	}
}
