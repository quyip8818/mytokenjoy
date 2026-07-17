package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAppendSyncLogPersists(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	entry := types.SyncLog{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000005901"), Time: "2026-06-10 10:00",
		Type: types.SyncTypeManual, Result: types.SyncResultSuccess, Detail: "ok",
	}
	if err := st.Org().AppendSyncLog(ctx, entry); err != nil {
		t.Fatal(err)
	}
	logs, err := st.Org().SyncLogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, log := range logs {
		if log.ID == entry.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected persisted sync log")
	}
}

func TestUsageBucketQuerySeriesHour(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	bucket := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart: bucket, DepartmentID: uuid.MustParse("00000000-0000-7000-0000-000000ddd001"), MemberID: uuid.MustParse("00000000-0000-7000-0000-000000eee001"),
		Model: "gpt-4o", Cost: 9000, DisplayCost: 9, CallCount: 2,
	}); err != nil {
		t.Fatal(err)
	}
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity:  types.UsageGranularityHour,
		Start:        time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		End:          time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
		GroupBy:      types.UsageGroupByNone,
		Timezone:     types.UsageDefaultTimezone,
		DepartmentID: uuid.MustParse("00000000-0000-7000-0000-000000ddd001"),
		MemberID:     uuid.MustParse("00000000-0000-7000-0000-000000eee001"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].Cost != 9 {
		t.Fatalf("expected one hour point with cost 9, got %+v", points)
	}
}
