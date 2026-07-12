package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgLedgerRepo) ListCallSettledAfterCursor(ctx context.Context, cursor store.LedgerProjectorCursor) ([]types.UsageLedgerEntry, error) {
	companyID := store.CompanyID(ctx)
	limit := cursor.Limit
	if limit <= 0 {
		limit = 500
	}

	var rows pgx.Rows
	var err error
	if cursor.LastOccurredAt == nil || cursor.LastLedgerID == nil {
		rows, err = r.db.Query(ctx, `
			SELECT `+ledgerSelectColumns+`
			FROM usage_ledger
			WHERE company_id = $1 AND event_type = 'call_settled'
			ORDER BY occurred_at ASC, id ASC
			LIMIT $2
		`, companyID, limit)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT `+ledgerSelectColumns+`
			FROM usage_ledger
			WHERE company_id = $1 AND event_type = 'call_settled'
			  AND (occurred_at, id) > ($2, $3)
			ORDER BY occurred_at ASC, id ASC
			LIMIT $4
		`, companyID, cursor.LastOccurredAt.UTC(), *cursor.LastLedgerID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLedgerRows(rows)
}

func (r *pgLedgerRepo) ListCallSettledSince(ctx context.Context, since time.Time) ([]types.UsageLedgerEntry, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT `+ledgerSelectColumns+`
		FROM usage_ledger
		WHERE company_id = $1 AND event_type = 'call_settled' AND occurred_at >= $2
		ORDER BY occurred_at ASC, id ASC
	`, companyID, since.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLedgerRows(rows)
}
