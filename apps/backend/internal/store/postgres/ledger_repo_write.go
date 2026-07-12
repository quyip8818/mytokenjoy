package postgres

import (
	"context"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

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
			department_id, member_id, project_id, platform_key_id, platform_key_scope,
			source, occurred_at, period_key, model, input_tokens, output_tokens,
			call_detail, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		ON CONFLICT (company_id, idempotency_key, lot_id, occurred_at) DO NOTHING
	`, entry.ID, companyID, entry.EventType, entry.IdempotencyKey, entry.SegmentIndex, entry.LotID,
		entry.Amount, entry.DisplayAmount, entry.BillingCurrency,
		entry.DepartmentID, entry.MemberID, entry.ProjectID, entry.PlatformKeyID, entry.PlatformKeyScope,
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
