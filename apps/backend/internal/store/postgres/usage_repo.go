package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/usagequery"
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
		points[i] = usagequery.AggregateToSeriesPoint(row)
	}
	usagequery.SortSeriesPoints(points)
	return points, nil
}

func (r *usageRepo) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	return r.queryAggregated(ctx, q)
}

func (r *usageRepo) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	rows, err := r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.DepartmentIDs, q.ScopeDeptIDs)
	if err != nil {
		return types.UsageSummaryTotals{}, err
	}
	return usagequery.SummaryTotals(rows, q.Start, q.End), nil
}

func (r *usageRepo) queryAggregated(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	rows, err := r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.DepartmentIDs, q.ScopeDeptIDs)
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

func (r *usageRepo) fetchFilteredRows(
	ctx context.Context,
	start, end time.Time,
	departmentID, memberID string,
	departmentIDs, scopeDeptIDs []string,
) ([]types.UsageBucketRow, error) {
	where, args := buildUsageWhere(start, end, departmentID, memberID, departmentIDs, scopeDeptIDs)
	query := fmt.Sprintf(`
		SELECT bucket_start, department_id, member_id, model,
			cost_cny, call_count, input_tokens, output_tokens
		FROM usage_buckets
		WHERE %s
	`, where)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]types.UsageBucketRow, 0)
	for rows.Next() {
		var row types.UsageBucketRow
		if err := rows.Scan(
			&row.BucketStart, &row.DepartmentID, &row.MemberID, &row.Model,
			&row.CostCNY, &row.CallCount, &row.InputTokens, &row.OutputTokens,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
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

var _ store.UsageRepository = (*usageRepo)(nil)
