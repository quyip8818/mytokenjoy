package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/usagequery"
)

type pgLedgerRepo struct {
	db dbQuerier
}

func (r *pgLedgerRepo) InsertOnConflict(ctx context.Context, entry types.UsageLedgerEntry) (bool, error) {
	return r.insertLedgerEntry(ctx, store.CompanyID(ctx), entry)
}

func (r *pgLedgerRepo) InsertSegments(ctx context.Context, entries []types.UsageLedgerEntry) (int, error) {
	companyID := store.CompanyID(ctx)
	inserted := 0
	for _, entry := range entries {
		ok, err := r.insertLedgerEntry(ctx, companyID, entry)
		if err != nil {
			return inserted, err
		}
		if ok {
			inserted++
		}
	}
	return inserted, nil
}

func (r *pgLedgerRepo) insertLedgerEntry(ctx context.Context, companyID int64, entry types.UsageLedgerEntry) (bool, error) {
	detailJSON, err := json.Marshal(entry.CallDetail)
	if err != nil {
		return false, err
	}
	tag, err := r.db.Exec(ctx, `
		INSERT INTO usage_ledger (
			id, company_id, event_type, idempotency_key, segment_index, lot_id,
			amount, display_amount, billing_currency,
			department_id, member_id, budget_group_id, platform_key_id,
			source, occurred_at, period_key, model, input_tokens, output_tokens,
			call_detail, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (company_id, idempotency_key, lot_id, occurred_at) DO NOTHING
	`, entry.ID, companyID, entry.EventType, entry.IdempotencyKey, entry.SegmentIndex, entry.LotID,
		entry.Amount, entry.DisplayAmount, entry.BillingCurrency,
		entry.DepartmentID, entry.MemberID, entry.BudgetGroupID, entry.PlatformKeyID,
		entry.Source, entry.OccurredAt.UTC(), entry.PeriodKey, entry.Model, entry.InputTokens, entry.OutputTokens,
		detailJSON, entry.CreatedAt.UTC())
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *pgLedgerRepo) ExistsIdempotency(ctx context.Context, idempotencyKey string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM usage_ledger
			WHERE company_id = $1 AND idempotency_key = $2
			LIMIT 1
		)
	`, store.CompanyID(ctx), idempotencyKey).Scan(&exists)
	return exists, err
}

func (r *pgLedgerRepo) ListCallSettledPage(ctx context.Context, filter store.LedgerCallFilter) ([]types.UsageLedgerEntry, int, error) {
	companyID := store.CompanyID(ctx)
	where, args := buildLedgerCallWhere(companyID, filter)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM usage_ledger WHERE %s
	`, where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizeLedgerPage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	listArgs := append(append([]any{}, args...), pageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, event_type, idempotency_key, segment_index, lot_id,
			amount, display_amount, billing_currency,
			department_id, member_id, budget_group_id, platform_key_id,
			source, occurred_at, period_key, model, input_tokens, output_tokens,
			call_detail, created_at
		FROM usage_ledger
		WHERE %s
		ORDER BY occurred_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanLedgerRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgLedgerRepo) QueryMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	companyID := store.CompanyID(ctx)
	where, args := buildLedgerSeriesWhere(companyID, q)
	query := fmt.Sprintf(`
		SELECT occurred_at, department_id, COALESCE(member_id, ''), model,
			amount, input_tokens, output_tokens
		FROM usage_ledger
		WHERE %s
	`, where)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loc, err := common.LoadLocation(q.Timezone)
	if err != nil {
		return nil, err
	}
	bucketRows := make([]types.UsageBucketRow, 0)
	for rows.Next() {
		var occurredAt time.Time
		var row types.UsageBucketRow
		var amount float64
		var inputTokens, outputTokens int64
		if err := rows.Scan(
			&occurredAt, &row.DepartmentID, &row.MemberID, &row.Model,
			&amount, &inputTokens, &outputTokens,
		); err != nil {
			return nil, err
		}
		row.BucketStart = usagequery.TruncateBucket(occurredAt, types.UsageGranularityMinute, loc)
		row.Cost = amount
		row.CallCount = 1
		row.InputTokens = inputTokens
		row.OutputTokens = outputTokens
		bucketRows = append(bucketRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	aggregated := usagequery.AggregateRows(bucketRows, types.UsageGranularityMinute, q.GroupBy, loc)
	points := make([]types.UsageSeriesPoint, len(aggregated))
	for i, row := range aggregated {
		points[i] = usagequery.AggregateToSeriesPoint(row)
	}
	usagequery.SortSeriesPoints(points)
	return points, nil
}

func buildLedgerCallWhere(companyID int64, filter store.LedgerCallFilter) (string, []any) {
	clauses := []string{"company_id = $1", "event_type = $2"}
	args := []any{companyID, types.EventTypeCallSettled}
	idx := 3

	if filter.Model != "" {
		clauses = append(clauses, fmt.Sprintf("model = $%d", idx))
		args = append(args, filter.Model)
		idx++
	}
	if filter.Status != "" {
		clauses = append(clauses, fmt.Sprintf("call_detail->>'status' = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.CallerID != "" {
		clauses = append(clauses, fmt.Sprintf("call_detail->>'callerId' = $%d", idx))
		args = append(args, filter.CallerID)
		idx++
	}
	if filter.From != "" {
		clauses = append(clauses, fmt.Sprintf("occurred_at::date >= $%d::date", idx))
		args = append(args, filter.From)
		idx++
	}
	if filter.To != "" {
		clauses = append(clauses, fmt.Sprintf("occurred_at::date <= $%d::date", idx))
		args = append(args, filter.To)
		idx++
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		clauses = append(clauses, fmt.Sprintf(`(
			model ILIKE $%d OR
			call_detail->>'caller' ILIKE $%d OR
			call_detail->>'previewSnippet' ILIKE $%d
		)`, idx, idx, idx))
		args = append(args, "%"+escapeLikePattern(kw)+"%")
		idx++
	}
	return strings.Join(clauses, " AND "), args
}

func buildLedgerSeriesWhere(companyID int64, q types.UsageSeriesQuery) (string, []any) {
	clauses := []string{"company_id = $1", "event_type = $2", "occurred_at >= $3", "occurred_at < $4"}
	args := []any{companyID, types.EventTypeCallSettled, q.Start.UTC(), q.End.UTC()}
	idx := 5
	if q.DepartmentID != "" {
		clauses = append(clauses, fmt.Sprintf("department_id = $%d", idx))
		args = append(args, q.DepartmentID)
		idx++
	}
	if q.MemberID != "" {
		clauses = append(clauses, fmt.Sprintf("COALESCE(member_id, '') = $%d", idx))
		args = append(args, q.MemberID)
		idx++
	}
	if len(q.ScopeDeptIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf("department_id = ANY($%d)", idx))
		args = append(args, q.ScopeDeptIDs)
		idx++
	}
	return strings.Join(clauses, " AND "), args
}

func normalizeLedgerPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

func scanLedgerRows(rows pgx.Rows) ([]types.UsageLedgerEntry, error) {
	items := make([]types.UsageLedgerEntry, 0)
	for rows.Next() {
		var item types.UsageLedgerEntry
		var memberID, budgetGroupID *string
		var detailJSON []byte
		var occurredAt, createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.EventType, &item.IdempotencyKey, &item.SegmentIndex, &item.LotID,
			&item.Amount, &item.DisplayAmount, &item.BillingCurrency,
			&item.DepartmentID, &memberID, &budgetGroupID, &item.PlatformKeyID,
			&item.Source, &occurredAt, &item.PeriodKey, &item.Model, &item.InputTokens, &item.OutputTokens,
			&detailJSON, &createdAt,
		); err != nil {
			return nil, err
		}
		item.MemberID = memberID
		item.BudgetGroupID = budgetGroupID
		item.OccurredAt = occurredAt
		item.CreatedAt = createdAt
		if len(detailJSON) > 0 {
			if err := json.Unmarshal(detailJSON, &item.CallDetail); err != nil {
				return nil, fmt.Errorf("unmarshal call_detail: %w", err)
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

var _ store.LedgerRepository = (*pgLedgerRepo)(nil)
