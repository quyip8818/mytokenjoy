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
	for _, row := range buildUsageBuckets(cfg.SeedReferenceDate()) {
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

func buildUsageBuckets(refDate string) []types.UsageBucketRow {
	anchor, err := time.Parse("2006-01-02", refDate)
	if err != nil {
		anchor = time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)
	}
	currentMonth := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := currentMonth.AddDate(0, -1, 0)

	rootConsumed := contract.DemoRootConsumed()
	const rawTotal = 39.5
	scale := rootConsumed / rawTotal

	// Richer demo data: multiple days, departments, members, and models.
	type entry struct {
		day    int
		hour   int
		dept   string
		member string
		model  string
		cost   float64
		calls  int
	}

	lastMonthEntries := []entry{
		{5, 9, contract.IDDept3, contract.IDMember1, "gpt-4o", 1.2, 32},
		{8, 10, contract.IDDept3, contract.IDMember1, "gpt-4o", 2.0, 45},
		{10, 11, contract.IDDept4, "m-4", "claude-3-5-sonnet", 1.8, 28},
		{12, 14, contract.IDDept3, contract.IDMemberPure, "gpt-4o-mini", 0.9, 67},
		{15, 8, contract.IDDept3, contract.IDMember1, "gpt-4o", 5.0, 85},
		{18, 10, contract.IDDept4, "m-4", "gpt-4o", 2.5, 41},
		{20, 15, contract.IDDept5, contract.IDMember3, "claude-3-5-sonnet", 1.5, 22},
		{22, 9, contract.IDDept3, contract.IDMember1, "gpt-4o-mini", 1.0, 88},
		{25, 11, contract.IDDept4, "m-4", "gpt-4o", 3.2, 56},
		{27, 16, contract.IDDept5, contract.IDMember3, "gpt-4o", 1.4, 31},
	}

	// Current month: richer daily distribution up to anchor day
	maxDay := anchor.Day()
	currentMonthEntries := []entry{
		{1, 9, contract.IDDept3, contract.IDMember1, "gpt-4o", 2.8, 52},
		{1, 14, contract.IDDept4, "m-4", "claude-3-5-sonnet", 1.5, 24},
		{2, 10, contract.IDDept3, contract.IDMemberPure, "gpt-4o-mini", 1.2, 95},
		{2, 15, contract.IDDept5, contract.IDMember3, "gpt-4o", 0.8, 19},
		{3, 9, contract.IDDept3, contract.IDMember1, "gpt-4o", 3.5, 68},
		{3, 11, contract.IDDept4, "m-4", "gpt-4o", 2.2, 37},
		{4, 10, contract.IDDept3, contract.IDMember1, "claude-3-5-sonnet", 4.1, 45},
		{4, 14, contract.IDDept5, contract.IDMember3, "gpt-4o-mini", 0.6, 48},
		{5, 9, contract.IDDept3, contract.IDMemberPure, "gpt-4o", 2.9, 61},
		{5, 11, contract.IDDept4, "m-4", "claude-3-5-sonnet", 3.8, 42},
		{5, 16, contract.IDDept3, contract.IDMember1, "gpt-4o-mini", 1.0, 82},
		{6, 10, contract.IDDept3, contract.IDMember1, "gpt-4o", 5.2, 98},
		{6, 14, contract.IDDept4, "m-4", "gpt-4o", 2.0, 33},
		{6, 16, contract.IDDept5, contract.IDMember3, "claude-3-5-sonnet", 1.3, 20},
		{7, 9, contract.IDDept3, contract.IDMember1, "gpt-4o", 4.6, 87},
		{7, 11, contract.IDDept3, contract.IDMemberPure, "gpt-4o-mini", 1.8, 112},
		{7, 15, contract.IDDept4, "m-4", "gpt-4o", 3.0, 49},
		{8, 10, contract.IDDept3, contract.IDMember1, "gpt-4o", 5.8, 105},
		{8, 14, contract.IDDept4, "m-4", "claude-3-5-sonnet", 2.6, 38},
		{8, 16, contract.IDDept5, contract.IDMember3, "gpt-4o", 1.9, 29},
		{9, 9, contract.IDDept3, contract.IDMember1, "gpt-4o", 3.2, 64},
		{9, 11, contract.IDDept3, contract.IDMemberPure, "gpt-4o", 2.1, 43},
	}

	var rows []types.UsageBucketRow

	for _, e := range lastMonthEntries {
		rows = append(rows, types.UsageBucketRow{
			BucketStart:  time.Date(lastMonth.Year(), lastMonth.Month(), e.day, e.hour, 0, 0, 0, time.UTC),
			DepartmentID: e.dept, MemberID: e.member, Model: e.model,
			Cost: e.cost * scale, DisplayCost: e.cost * scale, CallCount: e.calls,
		})
	}

	for _, e := range currentMonthEntries {
		if e.day > maxDay {
			continue
		}
		rows = append(rows, types.UsageBucketRow{
			BucketStart:  time.Date(currentMonth.Year(), currentMonth.Month(), e.day, e.hour, 0, 0, 0, time.UTC),
			DepartmentID: e.dept, MemberID: e.member, Model: e.model,
			Cost: e.cost * scale, DisplayCost: e.cost * scale, CallCount: e.calls,
		})
	}

	return rows
}
