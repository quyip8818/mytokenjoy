package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUsageBucketQuerySeriesHour(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := context.Background()
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{CostCNY: 3})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		BucketStart: time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC),
		CostCNY: 7,
	})
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 {
		t.Fatalf("expected two hour points, got %+v", points)
	}
}
