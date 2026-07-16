//go:build testhook

package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func EnsureBootstrapCompanyForTest(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	return ensureBootstrapCompany(ctx, pool, cfg)
}

// Usage aggregate test hooks (pure functions moved from store/usagequery).

func TestHookContainsString(items []string, target string) bool {
	return containsString(items, target)
}

func TestHookTruncateUsageBucket(t time.Time, granularity string, loc *time.Location) time.Time {
	return truncateUsageBucket(t, granularity, loc)
}

func TestHookSummaryUsageTotals(rows []types.UsageBucketRow, start, end time.Time) types.UsageSummaryTotals {
	return summaryUsageTotals(rows, start, end)
}

func TestHookLimitUsageByCost(rows []types.UsageAggregateRow, limit int) []types.UsageAggregateRow {
	return limitUsageByCost(rows, limit)
}

func TestHookTopModelPerDepartment(rows []types.UsageBucketRow, deptIDs []string) map[string]string {
	return topModelPerDepartment(rows, deptIDs)
}

func TestHookAggregateUsageRows(rows []types.UsageBucketRow, granularity, groupBy string, loc *time.Location) []types.UsageAggregateRow {
	return aggregateUsageRows(rows, granularity, groupBy, loc)
}

func TestHookSortUsageSeriesPoints(points []types.UsageSeriesPoint) {
	sortUsageSeriesPoints(points)
}
