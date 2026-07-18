package postgres

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type usageAggKey struct {
	Bucket       string
	DepartmentID uuid.UUID
	MemberID     uuid.UUID
	Model        string
}

func containsUUID(items []uuid.UUID, target uuid.UUID) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func truncateUsageBucket(t time.Time, granularity string, loc *time.Location) time.Time {
	switch granularity {
	case types.UsageGranularityDay:
		return common.TruncateInTZ(t, 24*time.Hour, loc)
	case types.UsageGranularityHour:
		return common.TruncateInTZ(t, time.Hour, loc)
	case types.UsageGranularityMinute:
		return common.TruncateInTZ(t, time.Minute, loc)
	case types.UsageGranularityWeek:
		return common.TruncateWeekInTZ(t, loc)
	case types.UsageGranularityMonth:
		return common.TruncateMonthInTZ(t, loc)
	default:
		return common.TruncateInTZ(t, 24*time.Hour, loc)
	}
}

func aggregateUsageRows(
	rows []types.UsageBucketRow,
	granularity, groupBy string,
	loc *time.Location,
) []types.UsageAggregateRow {
	if groupBy == "" {
		groupBy = types.UsageGroupByNone
	}
	buckets := make(map[usageAggKey]types.UsageAggregateRow)
	for _, row := range rows {
		var truncated time.Time
		var bucketLabel string
		if granularity != "" {
			truncated = truncateUsageBucket(row.BucketStart, granularity, loc)
			bucketLabel = truncated.UTC().Format(time.RFC3339)
		}
		key := usageAggKey{Bucket: bucketLabel}
		switch groupBy {
		case types.UsageGroupByDepartment:
			key.DepartmentID = row.DepartmentID
		case types.UsageGroupByMember:
			key.MemberID = row.MemberID
		case types.UsageGroupByModel:
			key.Model = row.Model
		case types.UsageGroupByNone:
			if granularity == "" {
				key.Bucket = ""
			}
		}
		existing := buckets[key]
		if existing.Bucket == "" && granularity != "" {
			existing.Bucket = common.FormatBucketISO(truncated)
		}
		existing.DepartmentID = key.DepartmentID
		existing.MemberID = key.MemberID
		existing.Model = key.Model
		existing.Cost += row.Cost
		existing.DisplayCost += row.DisplayCost
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

func aggregateToSeriesPoint(row types.UsageAggregateRow) types.UsageSeriesPoint {
	return types.UsageSeriesPoint{
		Bucket:       row.Bucket,
		DepartmentID: row.DepartmentID,
		MemberID:     row.MemberID,
		Model:        row.Model,
		Cost:         row.Spend(),
		CallCount:    row.CallCount,
		InputTokens:  row.InputTokens,
		OutputTokens: row.OutputTokens,
	}
}

func sortUsageSeriesPoints(points []types.UsageSeriesPoint) {
	sort.Slice(points, func(i, j int) bool {
		if points[i].Bucket != points[j].Bucket {
			return points[i].Bucket < points[j].Bucket
		}
		if points[i].DepartmentID != points[j].DepartmentID {
			return points[i].DepartmentID.String() < points[j].DepartmentID.String()
		}
		if points[i].MemberID != points[j].MemberID {
			return points[i].MemberID.String() < points[j].MemberID.String()
		}
		return points[i].Model < points[j].Model
	})
}

func summaryUsageTotals(rows []types.UsageBucketRow, start, end time.Time) types.UsageSummaryTotals {
	var totals types.UsageSummaryTotals
	for _, row := range rows {
		if row.BucketStart.Before(start) || !row.BucketStart.Before(end) {
			continue
		}
		totals.Cost += row.Cost
		totals.DisplayCost += row.DisplayCost
		totals.CallCount += row.CallCount
		totals.InputTokens += row.InputTokens
		totals.OutputTokens += row.OutputTokens
	}
	return totals
}

func limitUsageByCost(rows []types.UsageAggregateRow, limit int) []types.UsageAggregateRow {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Spend() > rows[j].Spend()
	})
	return rows[:limit]
}

func topModelPerDepartment(rows []types.UsageBucketRow, deptIDs []uuid.UUID) map[uuid.UUID]string {
	if len(deptIDs) == 0 {
		return map[uuid.UUID]string{}
	}
	deptSet := make(map[uuid.UUID]struct{}, len(deptIDs))
	for _, id := range deptIDs {
		deptSet[id] = struct{}{}
	}
	costs := make(map[uuid.UUID]map[string]float64)
	for _, row := range rows {
		if _, ok := deptSet[row.DepartmentID]; !ok {
			continue
		}
		if costs[row.DepartmentID] == nil {
			costs[row.DepartmentID] = make(map[string]float64)
		}
		costs[row.DepartmentID][row.Model] += row.Spend()
	}
	result := make(map[uuid.UUID]string, len(deptIDs))
	for deptID, models := range costs {
		topModel := ""
		topCost := 0.0
		for model, cost := range models {
			if cost > topCost {
				topCost = cost
				topModel = model
			}
		}
		result[deptID] = topModel
	}
	return result
}
