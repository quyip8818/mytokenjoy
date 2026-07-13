package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type projectionProgressRepo struct {
	db    dbQuerier
	table string
}

func newBudgetProjectionProgressRepo(db dbQuerier) store.ProjectionProgressRepository {
	return &projectionProgressRepo{db: db, table: "budget_projection_progress"}
}

func newDashboardProjectionProgressRepo(db dbQuerier) store.ProjectionProgressRepository {
	return &projectionProgressRepo{db: db, table: "dashboard_projection_progress"}
}

func (r *projectionProgressRepo) Get(ctx context.Context, stream string) (*store.ProjectionProgress, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT company_id, stream, last_occurred_at, last_ledger_id
		FROM %s
		WHERE company_id = $1 AND stream = $2
	`, r.table), companyID, stream)

	var progress store.ProjectionProgress
	err := row.Scan(&progress.CompanyID, &progress.Stream, &progress.LastOccurredAt, &progress.LastLedgerID)
	if err == pgx.ErrNoRows {
		return &store.ProjectionProgress{CompanyID: companyID, Stream: stream}, nil
	}
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *projectionProgressRepo) GetForUpdate(ctx context.Context, stream string) (*store.ProjectionProgress, error) {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (company_id, stream)
		VALUES ($1, $2)
		ON CONFLICT (company_id, stream) DO NOTHING
	`, r.table), companyID, stream); err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT company_id, stream, last_occurred_at, last_ledger_id
		FROM %s
		WHERE company_id = $1 AND stream = $2
		FOR UPDATE
	`, r.table), companyID, stream)

	var progress store.ProjectionProgress
	err := row.Scan(&progress.CompanyID, &progress.Stream, &progress.LastOccurredAt, &progress.LastLedgerID)
	if err == pgx.ErrNoRows {
		return &store.ProjectionProgress{CompanyID: companyID, Stream: stream}, nil
	}
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *projectionProgressRepo) Advance(ctx context.Context, stream string, lastOccurredAt time.Time, lastLedgerID string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (company_id, stream, last_occurred_at, last_ledger_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (company_id, stream) DO UPDATE SET
			last_occurred_at = EXCLUDED.last_occurred_at,
			last_ledger_id = EXCLUDED.last_ledger_id,
			updated_at = NOW()
	`, r.table), companyID, stream, lastOccurredAt.UTC(), lastLedgerID)
	return err
}
