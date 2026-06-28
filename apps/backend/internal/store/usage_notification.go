package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/timeutil"
)

type memoryUsageRepo struct {
	store *Memory
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
	loc, err := timeutil.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	aggregated := aggregateUsageRows(rows, q.Granularity, q.GroupBy, loc)
	points := make([]types.UsageSeriesPoint, 0, len(aggregated))
	for _, row := range aggregated {
		points = append(points, usageAggregateToSeriesPoint(row))
	}
	sortSeriesPoints(points)
	return points, nil
}

func (r *memoryUsageRepo) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	rows, err := r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, q.DepartmentIDs)
	if err != nil {
		return nil, err
	}
	loc, err := timeutil.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	aggregated := aggregateUsageRows(rows, q.Granularity, q.GroupBy, loc)
	if len(q.DepartmentIDs) > 0 {
		filtered := make([]types.UsageAggregateRow, 0)
		for _, row := range aggregated {
			if containsString(q.DepartmentIDs, row.DepartmentID) {
				filtered = append(filtered, row)
			}
		}
		aggregated = filtered
	}
	if q.Limit > 0 && len(aggregated) > q.Limit {
		sort.Slice(aggregated, func(i, j int) bool {
			return aggregated[i].CostCNY > aggregated[j].CostCNY
		})
		aggregated = aggregated[:q.Limit]
	}
	return aggregated, nil
}

func (r *memoryUsageRepo) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	rows, err := r.queryFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.ScopeDeptIDs, q.DepartmentIDs)
	if err != nil {
		return types.UsageSummaryTotals{}, err
	}
	var totals types.UsageSummaryTotals
	for _, row := range rows {
		if row.BucketStart.Before(q.Start) || !row.BucketStart.Before(q.End) {
			continue
		}
		totals.CostCNY += row.CostCNY
		totals.CallCount += row.CallCount
		totals.InputTokens += row.InputTokens
		totals.OutputTokens += row.OutputTokens
	}
	return totals, nil
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
	result := make([]types.UsageBucketRow, 0, len(r.store.usageBuckets))
	for _, row := range r.store.usageBuckets {
		if row.BucketStart.Before(start) || !row.BucketStart.Before(end) {
			continue
		}
		if departmentID != "" && row.DepartmentID != departmentID {
			continue
		}
		if len(departmentIDs) > 0 && !containsString(departmentIDs, row.DepartmentID) {
			continue
		}
		if memberID != "" && row.MemberID != memberID {
			continue
		}
		if len(scopeDeptIDs) > 0 && !containsString(scopeDeptIDs, row.DepartmentID) {
			continue
		}
		result = append(result, row)
	}
	return result, nil
}

type aggKey struct {
	bucket       string
	departmentID string
	memberID     string
	model        string
}

func aggregateUsageRows(
	rows []types.UsageBucketRow,
	granularity, groupBy string,
	loc *time.Location,
) []types.UsageAggregateRow {
	if groupBy == "" {
		groupBy = types.UsageGroupByNone
	}
	buckets := make(map[aggKey]types.UsageAggregateRow)
	for _, row := range rows {
		var truncated time.Time
		var bucketLabel string
		if granularity != "" {
			truncated = truncateBucket(row.BucketStart, granularity, loc)
			bucketLabel = truncated.UTC().Format(time.RFC3339)
		}
		key := aggKey{bucket: bucketLabel}
		switch groupBy {
		case types.UsageGroupByDepartment:
			key.departmentID = row.DepartmentID
		case types.UsageGroupByMember:
			key.memberID = row.MemberID
		case types.UsageGroupByModel:
			key.model = row.Model
		case types.UsageGroupByNone:
			if granularity == "" {
				key.bucket = ""
			}
		}
		existing := buckets[key]
		if existing.Bucket == "" && granularity != "" {
			existing.Bucket = timeutil.FormatBucketISO(truncated)
		}
		existing.DepartmentID = key.departmentID
		existing.MemberID = key.memberID
		existing.Model = key.model
		existing.CostCNY += row.CostCNY
		existing.CallCount += row.CallCount
		existing.InputTokens += row.InputTokens
		existing.OutputTokens += row.OutputTokens
		buckets[key] = existing
	}
	result := make([]types.UsageAggregateRow, 0, len(buckets))
	for _, row := range buckets {
		result = append(result, row)
	}
	return result
}

func truncateBucket(t time.Time, granularity string, loc *time.Location) time.Time {
	switch granularity {
	case types.UsageGranularityDay:
		return timeutil.TruncateInTZ(t, 24*time.Hour, loc)
	case types.UsageGranularityHour:
		return timeutil.TruncateInTZ(t, time.Hour, loc)
	case types.UsageGranularityWeek:
		return timeutil.TruncateWeekInTZ(t, loc)
	case types.UsageGranularityMonth:
		return timeutil.TruncateMonthInTZ(t, loc)
	default:
		return timeutil.TruncateInTZ(t, 24*time.Hour, loc)
	}
}

func usageAggregateToSeriesPoint(row types.UsageAggregateRow) types.UsageSeriesPoint {
	return types.UsageSeriesPoint{
		Bucket:       row.Bucket,
		DepartmentID: row.DepartmentID,
		MemberID:     row.MemberID,
		Model:        row.Model,
		CostCNY:      row.CostCNY,
		CallCount:    row.CallCount,
		InputTokens:  row.InputTokens,
		OutputTokens: row.OutputTokens,
	}
}

func sortSeriesPoints(points []types.UsageSeriesPoint) {
	sort.Slice(points, func(i, j int) bool {
		if points[i].Bucket != points[j].Bucket {
			return points[i].Bucket < points[j].Bucket
		}
		if points[i].DepartmentID != points[j].DepartmentID {
			return points[i].DepartmentID < points[j].DepartmentID
		}
		if points[i].MemberID != points[j].MemberID {
			return points[i].MemberID < points[j].MemberID
		}
		return points[i].Model < points[j].Model
	})
}

func childDepartmentIDs(departments []types.Department, parentID string) []string {
	parent := findDepartment(departments, parentID)
	if parent == nil {
		return nil
	}
	ids := make([]string, 0, len(parent.Children))
	for _, child := range parent.Children {
		ids = append(ids, child.ID)
	}
	return ids
}

func findDepartment(departments []types.Department, id string) *types.Department {
	for i := range departments {
		if departments[i].ID == id {
			return &departments[i]
		}
		if len(departments[i].Children) > 0 {
			if found := findDepartment(departments[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func (m *Memory) UsageBucketRows() []types.UsageBucketRow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]types.UsageBucketRow, 0, len(m.usageBuckets))
	for _, row := range m.usageBuckets {
		result = append(result, row)
	}
	return result
}

func (m *Memory) NotificationLogs() []types.NotificationLogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]types.NotificationLogEntry, len(m.notificationLogs))
	copy(result, m.notificationLogs)
	return result
}

var _ UsageRepository = (*memoryUsageRepo)(nil)

type memoryNotificationRepo struct {
	store *Memory
}

func (r *memoryNotificationRepo) Append(ctx context.Context, entry types.NotificationLogEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.notificationLogs = append(r.store.notificationLogs, entry)
	return nil
}

var _ NotificationRepository = (*memoryNotificationRepo)(nil)
