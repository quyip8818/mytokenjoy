package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) Upsert(ctx context.Context, nodeID uuid.UUID, row store.OrgNodeBudgetRow) error {
	return r.UpsertMany(ctx, []store.OrgNodeBudgetRow{{NodeID: nodeID, Budget: row.Budget, ReservedPool: row.ReservedPool, Period: row.Period, MemberAvgBudget: row.MemberAvgBudget}})
}

func (r *pgBudgetRepo) UpsertMany(ctx context.Context, rows []store.OrgNodeBudgetRow) error {
	if len(rows) == 0 {
		return nil
	}
	companyID := store.CompanyID(ctx)
	for _, row := range rows {
		period := row.Period
		if period == "" {
			period = pkgbudget.PeriodMonthly
		}
		if _, err := r.db.Exec(ctx, `
			UPDATE org_nodes SET
				budget = $3,
				reserved_pool = $4,
				period = $5,
				member_avg_budget = $6,
				updated_at = NOW()
			WHERE company_id = $1 AND id = $2
		`, companyID, row.NodeID, row.Budget, row.ReservedPool, period, row.MemberAvgBudget); err != nil {
			return fmt.Errorf("update org node budget %s: %w", row.NodeID, err)
		}
	}
	return nil
}

func (r *pgBudgetRepo) Get(ctx context.Context, nodeID uuid.UUID) (store.OrgNodeBudgetRow, bool, error) {
	companyID := store.CompanyID(ctx)
	var row store.OrgNodeBudgetRow
	row.NodeID = nodeID
	err := r.db.QueryRow(ctx, `
		SELECT budget, reserved_pool, period, member_avg_budget
		FROM org_nodes WHERE company_id = $1 AND id = $2
	`, companyID, nodeID).Scan(&row.Budget, &row.ReservedPool, &row.Period, &row.MemberAvgBudget)
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.OrgNodeBudgetRow{}, false, nil
		}
		return store.OrgNodeBudgetRow{}, false, err
	}
	return row, true, nil
}
