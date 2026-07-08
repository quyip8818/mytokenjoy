package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

type budgetSnapshotRepo struct {
	db dbQuerier
}

func newBudgetSnapshotRepo(db dbQuerier) *budgetSnapshotRepo {
	return &budgetSnapshotRepo{db: db}
}

var _ store.BudgetSnapshotRepository = (*budgetSnapshotRepo)(nil)

func (r *budgetSnapshotRepo) ListConsumed(ctx context.Context, axisKind, periodKey string) (map[string]float64, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT axis_id, consumed FROM budget_snapshots
		WHERE company_id = $1 AND axis_kind = $2 AND period_key = $3
	`, companyID, axisKind, periodKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]float64)
	for rows.Next() {
		var axisID string
		var consumed float64
		if err := rows.Scan(&axisID, &consumed); err != nil {
			return nil, err
		}
		out[axisID] = consumed
	}
	return out, rows.Err()
}

func (r *budgetSnapshotRepo) ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[string]float64, error) {
	if len(periodKeys) == 0 {
		return map[string]map[string]float64{}, nil
	}
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT period_key, axis_id, consumed FROM budget_snapshots
		WHERE company_id = $1 AND axis_kind = $2 AND period_key = ANY($3)
	`, companyID, axisKind, periodKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]map[string]float64)
	for rows.Next() {
		var periodKey, axisID string
		var consumed float64
		if err := rows.Scan(&periodKey, &axisID, &consumed); err != nil {
			return nil, err
		}
		if out[periodKey] == nil {
			out[periodKey] = make(map[string]float64)
		}
		out[periodKey][axisID] = consumed
	}
	return out, rows.Err()
}

func (r *budgetSnapshotRepo) GetConsumed(ctx context.Context, axisKind, axisID, periodKey string) (float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var consumed float64
	err := r.db.QueryRow(ctx, `
		SELECT consumed FROM budget_snapshots
		WHERE company_id = $1 AND axis_kind = $2 AND axis_id = $3 AND period_key = $4
	`, companyID, axisKind, axisID, periodKey).Scan(&consumed)
	if err == pgx.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return consumed, true, nil
}

func (r *budgetSnapshotRepo) IncrementConsumed(ctx context.Context, axisKind, axisID, periodKey string, amountCNY float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO budget_snapshots (company_id, axis_kind, axis_id, period_key, consumed, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (company_id, axis_kind, axis_id, period_key) DO UPDATE SET
			consumed = budget_snapshots.consumed + EXCLUDED.consumed,
			updated_at = NOW()
	`, companyID, axisKind, axisID, periodKey, amountCNY)
	return err
}

func (r *budgetSnapshotRepo) SetConsumed(ctx context.Context, axisKind, axisID, periodKey string, consumed float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO budget_snapshots (company_id, axis_kind, axis_id, period_key, consumed, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (company_id, axis_kind, axis_id, period_key) DO UPDATE SET
			consumed = EXCLUDED.consumed,
			updated_at = NOW()
	`, companyID, axisKind, axisID, periodKey, consumed)
	return err
}

func (r *budgetSnapshotRepo) RollupOrgNodeAncestors(ctx context.Context, leafNodeID, periodKey string, amountCNY float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO budget_snapshots (company_id, axis_kind, axis_id, period_key, consumed, updated_at)
		SELECT $1, $5, ancestor.id, $4, $3, NOW()
		FROM org_nodes leaf
		JOIN org_nodes ancestor
		  ON ancestor.company_id = leaf.company_id
		 AND ancestor.path @> leaf.path
		WHERE leaf.company_id = $1 AND leaf.id = $2
		ON CONFLICT (company_id, axis_kind, axis_id, period_key) DO UPDATE SET
			consumed = budget_snapshots.consumed + EXCLUDED.consumed,
			updated_at = NOW()
	`, companyID, leafNodeID, amountCNY, periodKey, store.SnapshotAxisOrgNode)
	return err
}
