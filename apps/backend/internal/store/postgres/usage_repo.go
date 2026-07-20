package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type usageRepo struct {
	db dbQuerier
}

const (
	usageBucketColumns = `bucket_start, department_id, member_id, model,
			quota_consumed, display_cost, call_count, input_tokens, output_tokens`
)

func (r *usageRepo) UpsertBucket(ctx context.Context, row types.UsageBucketRow) error {
	companyID := store.CompanyID(ctx)
	var memberID *uuid.UUID
	if row.MemberID != uuid.Nil {
		memberID = &row.MemberID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO usage_buckets (
			company_id, bucket_start, department_id, member_id, model,
			quota_consumed, display_cost, call_count, input_tokens, output_tokens, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (company_id, bucket_start, department_id, member_scope, model) DO UPDATE SET
			quota_consumed = usage_buckets.quota_consumed + EXCLUDED.quota_consumed,
			display_cost = usage_buckets.display_cost + EXCLUDED.display_cost,
			call_count = usage_buckets.call_count + EXCLUDED.call_count,
			input_tokens = usage_buckets.input_tokens + EXCLUDED.input_tokens,
			output_tokens = usage_buckets.output_tokens + EXCLUDED.output_tokens,
			updated_at = NOW()
	`, companyID, row.BucketStart.UTC(), row.DepartmentID, memberID, row.Model,
		row.QuotaConsumed, row.DisplayCost, row.CallCount, row.InputTokens, row.OutputTokens)
	return err
}

func (r *usageRepo) SetBucket(ctx context.Context, row types.UsageBucketRow) error {
	companyID := store.CompanyID(ctx)
	var memberID *uuid.UUID
	if row.MemberID != uuid.Nil {
		memberID = &row.MemberID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO usage_buckets (
			company_id, bucket_start, department_id, member_id, model,
			quota_consumed, display_cost, call_count, input_tokens, output_tokens, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (company_id, bucket_start, department_id, member_scope, model) DO UPDATE SET
			quota_consumed = EXCLUDED.quota_consumed,
			display_cost = EXCLUDED.display_cost,
			call_count = EXCLUDED.call_count,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			updated_at = NOW()
	`, companyID, row.BucketStart.UTC(), row.DepartmentID, memberID, row.Model,
		row.QuotaConsumed, row.DisplayCost, row.CallCount, row.InputTokens, row.OutputTokens)
	return err
}

func (r *usageRepo) ListBucketsSince(ctx context.Context, since time.Time) ([]types.UsageBucketRow, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT `+usageBucketColumns+`
		FROM usage_buckets
		WHERE company_id = $1 AND bucket_start >= $2
	`, companyID, since.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]types.UsageBucketRow, 0)
	for rows.Next() {
		var row types.UsageBucketRow
		var memberID *uuid.UUID
		if err := rows.Scan(
			&row.BucketStart, &row.DepartmentID, &memberID, &row.Model,
			&row.QuotaConsumed, &row.DisplayCost, &row.CallCount, &row.InputTokens, &row.OutputTokens,
		); err != nil {
			return nil, err
		}
		if memberID != nil {
			row.MemberID = *memberID
		}
		result = append(result, row)
	}
	return result, rows.Err()
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
		points[i] = aggregateToSeriesPoint(row)
	}
	sortUsageSeriesPoints(points)
	return points, nil
}

func (r *usageRepo) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	return r.queryAggregated(ctx, q)
}

func (r *usageRepo) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	rows, err := r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.OwnerDepartmentID, q.ScopeDeptIDs)
	if err != nil {
		return types.UsageSummaryTotals{}, err
	}
	return summaryUsageTotals(rows, q.Start, q.End), nil
}

func (r *usageRepo) QueryFilteredBuckets(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageBucketRow, error) {
	return r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.OwnerDepartmentID, q.ScopeDeptIDs)
}

func (r *usageRepo) TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.OwnerDepartmentID, q.ScopeDeptIDs)
	if err != nil {
		return nil, err
	}
	return topModelPerDepartment(rows, deptIDs), nil
}

func (r *usageRepo) queryAggregated(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	rows, err := r.fetchFilteredRows(ctx, q.Start, q.End, q.DepartmentID, q.MemberID, q.OwnerDepartmentID, q.ScopeDeptIDs)
	if err != nil {
		return nil, err
	}
	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	aggregated := aggregateUsageRows(rows, q.Granularity, q.GroupBy, loc)
	if len(q.OwnerDepartmentID) > 0 {
		ownerSet := uuidSet(q.OwnerDepartmentID)
		filtered := make([]types.UsageAggregateRow, 0)
		for _, row := range aggregated {
			if _, ok := ownerSet[row.DepartmentID]; ok {
				filtered = append(filtered, row)
			}
		}
		aggregated = filtered
	}
	return limitUsageByCost(aggregated, q.Limit), nil
}

func (r *usageRepo) fetchFilteredRows(
	ctx context.Context,
	start, end time.Time,
	departmentID, memberID uuid.UUID,
	departmentIDs, scopeDeptIDs []uuid.UUID,
) ([]types.UsageBucketRow, error) {
	companyID := store.CompanyID(ctx)
	where, args := buildUsageWhere(companyID, start, end, departmentID, memberID, departmentIDs, scopeDeptIDs)
	query := fmt.Sprintf(`
		SELECT `+usageBucketColumns+`
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
		var memberID *uuid.UUID
		if err := rows.Scan(
			&row.BucketStart, &row.DepartmentID, &memberID, &row.Model,
			&row.QuotaConsumed, &row.DisplayCost, &row.CallCount, &row.InputTokens, &row.OutputTokens,
		); err != nil {
			return nil, err
		}
		if memberID != nil {
			row.MemberID = *memberID
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func buildUsageWhere(
	companyID uuid.UUID,
	start, end time.Time,
	departmentID, memberID uuid.UUID,
	departmentIDs, scopeDeptIDs []uuid.UUID,
) (string, []any) {
	args := []any{companyID, start.UTC(), end.UTC()}
	clauses := []string{
		"company_id = $1",
		"bucket_start >= $2",
		"bucket_start < $3",
	}
	idx := 4
	if departmentID != uuid.Nil {
		clauses = append(clauses, fmt.Sprintf("department_id = $%d", idx))
		args = append(args, departmentID)
		idx++
	}
	if len(departmentIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf("department_id = ANY($%d)", idx))
		args = append(args, departmentIDs)
		idx++
	}
	if memberID != uuid.Nil {
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
