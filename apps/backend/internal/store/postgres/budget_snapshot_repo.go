package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type budgetSnapshotRepo struct {
	db dbQuerier
}

func newBudgetSnapshotRepo(db dbQuerier) *budgetSnapshotRepo {
	return &budgetSnapshotRepo{db: db}
}

func (r *budgetSnapshotRepo) Get(ctx context.Context, periodKey string) (*store.BudgetSnapshot, error) {
	companyID := store.CompanyID(ctx)
	var snapshot json.RawMessage
	err := r.db.QueryRow(ctx, `
		SELECT snapshot FROM budget_snapshot
		WHERE company_id = $1 AND period_key = $2
	`, companyID, periodKey).Scan(&snapshot)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &store.BudgetSnapshot{PeriodKey: periodKey, Snapshot: snapshot}, nil
}

func (r *budgetSnapshotRepo) GetForUpdate(ctx context.Context, periodKey string) (*store.BudgetSnapshot, error) {
	companyID := store.CompanyID(ctx)
	var snapshot json.RawMessage
	err := r.db.QueryRow(ctx, `
		SELECT snapshot FROM budget_snapshot
		WHERE company_id = $1 AND period_key = $2
		FOR UPDATE
	`, companyID, periodKey).Scan(&snapshot)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &store.BudgetSnapshot{PeriodKey: periodKey, Snapshot: snapshot}, nil
}

func (r *budgetSnapshotRepo) Upsert(ctx context.Context, periodKey string, snapshot json.RawMessage) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO budget_snapshot (company_id, period_key, snapshot)
		VALUES ($1, $2, $3)
		ON CONFLICT (company_id, period_key)
		DO UPDATE SET snapshot = EXCLUDED.snapshot, updated_at = NOW()
	`, companyID, periodKey, snapshot)
	return err
}

func (r *budgetSnapshotRepo) Delete(ctx context.Context, periodKey string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		DELETE FROM budget_snapshot WHERE company_id = $1 AND period_key = $2
	`, companyID, periodKey)
	return err
}

func (r *budgetSnapshotRepo) ListPeriods(ctx context.Context) ([]string, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT period_key FROM budget_snapshot
		WHERE company_id = $1
		ORDER BY period_key ASC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var periods []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		periods = append(periods, pk)
	}
	return periods, rows.Err()
}

func (r *budgetSnapshotRepo) Exists(ctx context.Context, periodKey string) (bool, error) {
	companyID := store.CompanyID(ctx)
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM budget_snapshot WHERE company_id = $1 AND period_key = $2)
	`, companyID, periodKey).Scan(&exists)
	return exists, err
}

var _ store.BudgetSnapshotRepository = (*budgetSnapshotRepo)(nil)
