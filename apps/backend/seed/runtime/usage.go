package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func ApplyUsageBuckets(ctx context.Context, st store.Store, cfg config.Config) error {
	if _, ok := company.FromContext(ctx); !ok {
		ctx = company.DefaultContext(contract.DefaultCompanyID)
	}
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
	return totals.CallCount == 0 && totals.Cost == 0, nil
}

func buildUsageBuckets(demoToday string) []types.UsageBucketRow {
	anchor, err := time.Parse("2006-01-02", demoToday)
	if err != nil {
		anchor = time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	}
	currentMonth := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := currentMonth.AddDate(0, -1, 0)

	rootConsumed := contract.DemoRootConsumed()
	const rawTotal = 39.5
	scale := rootConsumed / rawTotal

	return []types.UsageBucketRow{
		{
			BucketStart:  time.Date(lastMonth.Year(), lastMonth.Month(), 15, 8, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept3, MemberID: contract.IDMember1, Model: "gpt-4o",
			Cost: 5 * scale, CallCount: 85,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 10, 8, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept3, MemberID: contract.IDMember1, Model: "gpt-4o",
			Cost: 12.5 * scale, CallCount: 128,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 10, 9, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept3, MemberID: contract.IDMember1, Model: "gpt-4o-mini",
			Cost: 4 * scale, CallCount: 214,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 12, 10, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept4, MemberID: "m-4", Model: "claude-3-5-sonnet",
			Cost: 8.5 * scale, CallCount: 86,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 14, 11, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept3, MemberID: contract.IDMemberPure, Model: "gpt-4o",
			Cost: 6 * scale, CallCount: 171,
		},
		{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), 16, 14, 0, 0, 0, time.UTC),
			DepartmentID: contract.IDDept4, MemberID: "m-4", Model: "gpt-4o",
			Cost: 3.5 * scale, CallCount: 43,
		},
	}
}
