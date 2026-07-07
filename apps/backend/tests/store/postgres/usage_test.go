//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUsageBucketUpsertAccumulates(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC)
	row := types.UsageBucketRow{
		BucketStart:  bucket,
		DepartmentID: "dept-usage-test",
		MemberID:     "m-usage-test",
		Model:        "gpt-4o",
		CostCNY:      1.5,
		CallCount:    1,
	}
	if err := st.Usage().UpsertBucket(ctx, row); err != nil {
		t.Fatal(err)
	}
	row.CostCNY = 2.5
	row.CallCount = 1
	if err := st.Usage().UpsertBucket(ctx, row); err != nil {
		t.Fatal(err)
	}

	dbURL := getTestDB(t).url
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	var cost float64
	var callCount int
	err = conn.QueryRow(ctx, `
		SELECT cost_cny, call_count FROM usage_buckets
		WHERE bucket_start = $1 AND department_id = $2 AND member_id = $3 AND model = $4
	`, bucket, row.DepartmentID, row.MemberID, row.Model).Scan(&cost, &callCount)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 4.0 || callCount != 2 {
		t.Fatalf("expected accumulated usage 4.0/2 calls, got cost=%v calls=%d", cost, callCount)
	}
}

func TestUsageBucketQuerySeriesDay(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: "dept-series", MemberID: "m-1",
		Model: "gpt-4o", CostCNY: 5, CallCount: 2,
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
	if len(points) != 1 || points[0].CostCNY != 5 {
		t.Fatalf("expected one day point with cost 5, got %+v", points)
	}
}
