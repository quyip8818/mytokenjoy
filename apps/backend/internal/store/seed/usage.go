package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func ApplyUsageBuckets(ctx context.Context, st store.Store, cfg config.Config) error {
	empty, err := usageBucketsEmpty(ctx, st)
	if err != nil {
		return fmt.Errorf("check usage buckets: %w", err)
	}
	if !empty {
		return nil
	}
	for _, row := range buildUsageBuckets(cfg.DemoToday) {
		if err := st.Usage().UpsertBucket(ctx, row); err != nil {
			return fmt.Errorf("seed usage bucket: %w", err)
		}
	}
	return nil
}

func usageBucketsEmpty(ctx context.Context, st store.Store) (bool, error) {
	totals, err := st.Usage().QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:      time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		return false, err
	}
	return totals.CallCount == 0 && totals.CostCNY == 0, nil
}

func buildUsageBuckets(demoToday string) []types.UsageBucketRow {
	anchor, err := time.Parse("2006-01-02", demoToday)
	if err != nil {
		anchor = time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	}
	currentMonth := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := currentMonth.AddDate(0, -1, 0)

	return []types.UsageBucketRow{
		{
			BucketStart:  time.Date(lastMonth.Year(), lastMonth.Month(), 15, 8, 0, 0, 0, time.UTC),
			DepartmentID: IDDept3, MemberID: IDMember1, Model: "gpt-4o",
			CostCNY: 5, CallCount: 2,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 10, 8, 0, 0, 0, time.UTC),
			DepartmentID: IDDept3, MemberID: IDMember1, Model: "gpt-4o",
			CostCNY: 12.5, CallCount: 3,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 10, 9, 0, 0, 0, time.UTC),
			DepartmentID: IDDept3, MemberID: IDMember1, Model: "gpt-4o-mini",
			CostCNY: 4, CallCount: 5,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 12, 10, 0, 0, 0, time.UTC),
			DepartmentID: IDDept4, MemberID: "m-4", Model: "claude-3-5-sonnet",
			CostCNY: 8.5, CallCount: 2,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 14, 11, 0, 0, 0, time.UTC),
			DepartmentID: IDDept3, MemberID: "m-pure", Model: "gpt-4o",
			CostCNY: 6, CallCount: 4,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 16, 14, 0, 0, 0, time.UTC),
			DepartmentID: IDDept4, MemberID: "m-4", Model: "gpt-4o",
			CostCNY: 3.5, CallCount: 1,
		},
	}
}
