package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/usagequery"
)

type memoryUsageRepo struct {
	store *Store
}

func usageBucketKey(row types.UsageBucketRow) string {
	return fmt.Sprintf("%s|%s|%s|%s",
		row.BucketStart.UTC().Format(time.RFC3339),
		row.DepartmentID, row.MemberID, row.Model,
	)
}

func (r *memoryUsageRepo) UpsertBucket(ctx context.Context, row types.UsageBucketRow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.usageBuckets == nil {
		r.store.usageBuckets = make(map[string]types.UsageBucketRow)
	}
	key := usageBucketKey(row)
	existing, ok := r.store.usageBuckets[key]
	if !ok {
		row.BucketStart = row.BucketStart.UTC()
		r.store.usageBuckets[key] = row
		return nil
	}
	existing.CostCNY += row.CostCNY
	existing.CallCount += row.CallCount
	existing.InputTokens += row.InputTokens
	existing.OutputTokens += row.OutputTokens
	r.store.usageBuckets[key] = existing
	return nil
}

func (r *memoryUsageRepo) QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	rows, err := r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, nil)
	if err != nil {
		return nil, err
	}
	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	aggregated := usagequery.AggregateRows(rows, q.Granularity, q.GroupBy, loc)
	points := make([]types.UsageSeriesPoint, 0, len(aggregated))
	for _, row := range aggregated {
		points = append(points, usagequery.AggregateToSeriesPoint(row))
	}
	usagequery.SortSeriesPoints(points)
	return points, nil
}

func (r *memoryUsageRepo) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	rows, err := r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, q.DepartmentIDs)
	if err != nil {
		return nil, err
	}
	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	aggregated := usagequery.AggregateRows(rows, q.Granularity, q.GroupBy, loc)
	if len(q.DepartmentIDs) > 0 {
		filtered := make([]types.UsageAggregateRow, 0)
		for _, row := range aggregated {
			if usagequery.ContainsString(q.DepartmentIDs, row.DepartmentID) {
				filtered = append(filtered, row)
			}
		}
		aggregated = filtered
	}
	return usagequery.LimitByCost(aggregated, q.Limit), nil
}

func (r *memoryUsageRepo) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	rows, err := r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, q.DepartmentIDs)
	if err != nil {
		return types.UsageSummaryTotals{}, err
	}
	return usagequery.SummaryTotals(rows, q.Start, q.End), nil
}

func (r *memoryUsageRepo) QueryFilteredBuckets(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageBucketRow, error) {
	return r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, q.DepartmentIDs)
}

func (r *memoryUsageRepo) queryFilteredRows(
	ctx context.Context,
	start, end time.Time,
	departmentID, memberID string,
	scopeDeptIDs, departmentIDs []string,
) ([]types.UsageBucketRow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	all := make([]types.UsageBucketRow, 0, len(r.store.usageBuckets))
	for _, row := range r.store.usageBuckets {
		all = append(all, row)
	}
	return usagequery.FilterRows(all, start, end, departmentID, memberID, scopeDeptIDs, departmentIDs), nil
}

func (m *Store) UsageBucketRows() []types.UsageBucketRow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]types.UsageBucketRow, 0, len(m.usageBuckets))
	for _, row := range m.usageBuckets {
		result = append(result, row)
	}
	return result
}

var _ store.UsageRepository = (*memoryUsageRepo)(nil)
