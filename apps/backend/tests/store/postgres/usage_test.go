package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUsageBucketUpsertAccumulates(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC)
	row := types.UsageBucketRow{
		BucketStart:   bucket,
		DepartmentID:  uuid.MustParse("00000000-0000-7000-8000-00000000dd01"),
		MemberID:      uuid.MustParse("00000000-0000-7000-8000-00000000ee01"),
		Model:         "gpt-4o",
		QuotaConsumed: 1,
		DisplayCost:   0.015,
		CallCount:     1,
	}
	if err := st.Usage().UpsertBucket(ctx, row); err != nil {
		t.Fatal(err)
	}
	row.QuotaConsumed = 2
	row.DisplayCost = 0.025
	row.CallCount = 1
	if err := st.Usage().UpsertBucket(ctx, row); err != nil {
		t.Fatal(err)
	}

	dbURL := testutil.TestSchemaURL(t)
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	var quotaConsumed int64
	var displayCost float64
	var callCount int
	err = conn.QueryRow(ctx, `
		SELECT quota_consumed, display_cost, call_count FROM usage_buckets
		WHERE bucket_start = $1 AND department_id = $2 AND member_id = $3 AND model = $4
	`, bucket, row.DepartmentID, row.MemberID, row.Model).Scan(&quotaConsumed, &displayCost, &callCount)
	if err != nil {
		t.Fatal(err)
	}
	if quotaConsumed != 3 || displayCost != 0.04 || callCount != 2 {
		t.Fatalf("expected accumulated usage 3/0.04/2 calls, got cost=%d display=%v calls=%d", quotaConsumed, displayCost, callCount)
	}
}

func TestUsageBucketQuerySeriesDay(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: uuid.MustParse("00000000-0000-7000-8000-00000000dd02"), MemberID: uuid.MustParse("00000000-0000-7000-8000-00000000ee02"),
		Model: "gpt-4o", QuotaConsumed: 5000, DisplayCost: 5, CallCount: 2,
	}); err != nil {
		t.Fatal(err)
	}
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityDay,
		Start:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].Cost != 5 {
		t.Fatalf("expected one day point with display cost 5, got %+v", points)
	}
}
