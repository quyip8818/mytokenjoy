package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type usageRepo struct {
	db dbQuerier
}

func (r *usageRepo) UpsertBucket(ctx context.Context, row types.UsageBucketRow) error {
	memberID := row.MemberID
	_, err := r.db.Exec(ctx, `
		INSERT INTO usage_buckets (
			bucket_start, department_id, member_id, model,
			cost_cny, call_count, input_tokens, output_tokens, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (bucket_start, department_id, member_id, model) DO UPDATE SET
			cost_cny = usage_buckets.cost_cny + EXCLUDED.cost_cny,
			call_count = usage_buckets.call_count + EXCLUDED.call_count,
			input_tokens = usage_buckets.input_tokens + EXCLUDED.input_tokens,
			output_tokens = usage_buckets.output_tokens + EXCLUDED.output_tokens,
			updated_at = NOW()
	`, row.BucketStart.UTC(), row.DepartmentID, memberID, row.Model,
		row.CostCNY, row.CallCount, row.InputTokens, row.OutputTokens)
	return err
}

func (r *usageRepo) QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	rows, err := r.queryAggregated(ctx, types.UsageAggregateQuery{
		Start:        q.Start,
		End:          q.End,
		Granularity:  q.Granularity,
		Timezone:     q.Timezone,
		GroupBy:      q.GroupBy,
		DepartmentID: q.DepartmentID,
		MemberID:     q.MemberID,
		ScopeDeptIDs: q.ScopeDeptIDs,
	})
	if err != nil {
		return nil, err
	}
	points := make([]types.UsageSeriesPoint, len(rows))
	for i, row := range rows {
		points[i] = types.UsageSeriesPoint{
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
	return points, nil
}

func (r *usageRepo) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	return r.queryAggregated(ctx, q)
}

func (r *usageRepo) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	where, args := buildUsageWhere(q.Start, q.End, q.DepartmentID, q.MemberID, q.DepartmentIDs, q.ScopeDeptIDs)
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(cost_cny), 0),
			COALESCE(SUM(call_count), 0),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0)
		FROM usage_buckets
		WHERE %s
	`, where)
	var totals types.UsageSummaryTotals
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&totals.CostCNY, &totals.CallCount, &totals.InputTokens, &totals.OutputTokens,
	)
	return totals, err
}

func (r *usageRepo) queryAggregated(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	groupBy := q.GroupBy
	if groupBy == "" {
		groupBy = types.UsageGroupByNone
	}
	where, args := buildUsageWhere(q.Start, q.End, q.DepartmentID, q.MemberID, q.DepartmentIDs, q.ScopeDeptIDs)

	var truncExpr string
	var selectBucket string
	var groupDims string
	if q.Granularity != "" {
		truncExpr = bucketTruncExpr(q.Granularity, q.Timezone)
		selectBucket = fmt.Sprintf("to_char(%s, 'YYYY-MM-DD\"T\"HH24:MI:SSOF') AS bucket,", truncExpr)
		groupDims = "bucket"
	} else {
		selectBucket = "'' AS bucket,"
		groupDims = ""
	}

	selectDims, dimGroup := aggregateSelectGroup(groupBy)
	if q.Granularity != "" {
		if dimGroup == "bucket" || dimGroup == "" {
			groupDims = "bucket"
		} else {
			groupDims = "bucket, " + dimGroup
		}
	} else {
		groupDims = dimGroup
		if groupDims == "bucket" {
			groupDims = "1"
			selectBucket = "'' AS bucket,"
		}
	}

	query := fmt.Sprintf(`
		SELECT
			%s
			%s
			COALESCE(SUM(cost_cny), 0) AS cost_cny,
			COALESCE(SUM(call_count), 0) AS call_count,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens
		FROM usage_buckets
		WHERE %s
		GROUP BY %s
	`, selectBucket, selectDims, where, groupDims)

	if q.Limit > 0 {
		query += fmt.Sprintf(" ORDER BY cost_cny DESC LIMIT %d", q.Limit)
	} else {
		query += " ORDER BY bucket ASC"
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]types.UsageAggregateRow, 0)
	for rows.Next() {
		var row types.UsageAggregateRow
		var dept, member, model *string
		scanTargets := []any{&row.Bucket}
		switch groupBy {
		case types.UsageGroupByDepartment:
			scanTargets = append(scanTargets, &dept)
		case types.UsageGroupByMember:
			scanTargets = append(scanTargets, &member)
		case types.UsageGroupByModel:
			scanTargets = append(scanTargets, &model)
		}
		scanTargets = append(scanTargets, &row.CostCNY, &row.CallCount, &row.InputTokens, &row.OutputTokens)
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}
		if dept != nil {
			row.DepartmentID = *dept
		}
		if member != nil {
			row.MemberID = *member
		}
		if model != nil {
			row.Model = *model
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func bucketTruncExpr(granularity, timezone string) string {
	tz := timezone
	if tz == "" {
		tz = types.UsageDefaultTimezone
	}
	local := fmt.Sprintf("(bucket_start AT TIME ZONE '%s')", tz)
	switch granularity {
	case types.UsageGranularityHour:
		return fmt.Sprintf("date_trunc('hour', %s)", local)
	case types.UsageGranularityWeek:
		return fmt.Sprintf("date_trunc('week', %s)", local)
	case types.UsageGranularityMonth:
		return fmt.Sprintf("date_trunc('month', %s)", local)
	default:
		return fmt.Sprintf("date_trunc('day', %s)", local)
	}
}

func aggregateSelectGroup(groupBy string) (selectCols, groupCols string) {
	switch groupBy {
	case types.UsageGroupByDepartment:
		return "department_id, ", "department_id"
	case types.UsageGroupByMember:
		return "member_id, ", "member_id"
	case types.UsageGroupByModel:
		return "model, ", "model"
	default:
		return "", "bucket"
	}
}

func buildUsageWhere(
	start, end time.Time,
	departmentID, memberID string,
	departmentIDs, scopeDeptIDs []string,
) (string, []any) {
	args := []any{start.UTC(), end.UTC()}
	clauses := []string{
		"bucket_start >= $1",
		"bucket_start < $2",
	}
	idx := 3
	if departmentID != "" {
		clauses = append(clauses, fmt.Sprintf("department_id = $%d", idx))
		args = append(args, departmentID)
		idx++
	}
	if len(departmentIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf("department_id = ANY($%d)", idx))
		args = append(args, departmentIDs)
		idx++
	}
	if memberID != "" {
		clauses = append(clauses, fmt.Sprintf("member_id = $%d", idx))
		args = append(args, memberID)
		idx++
	}
	if len(scopeDeptIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf("department_id = ANY($%d)", idx))
		args = append(args, scopeDeptIDs)
		idx++
	}
	return strings.Join(clauses, " AND "), args
}

type notificationRepo struct {
	db dbQuerier
}

func (r *notificationRepo) Append(ctx context.Context, entry types.NotificationLogEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_log (id, channel, event_type, recipient, payload, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NOW())
	`, entry.ID, entry.Channel, entry.EventType, entry.Recipient, entry.Payload, entry.Status, entry.Error)
	return err
}

var _ store.UsageRepository = (*usageRepo)(nil)
var _ store.NotificationRepository = (*notificationRepo)(nil)
