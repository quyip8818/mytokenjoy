package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) Tree(ctx context.Context) ([]types.BudgetNode, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, parent_id, budget, consumed, reserved_pool, period, sort_order
		FROM budget_nodes WHERE company_id = $1 ORDER BY sort_order
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flat := make([]flatBudgetNode, 0)
	for rows.Next() {
		var row flatBudgetNode
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.Budget, &row.Consumed,
			&row.ReservedPool, &row.Period, &row.sortOrder,
		); err != nil {
			return nil, err
		}
		flat = append(flat, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneBudgetTree(buildBudgetTree(flat)), nil
}

func (r *pgBudgetRepo) SetTree(ctx context.Context, tree []types.BudgetNode) error {
	companyID := store.CompanyID(ctx)
	flat := flattenBudgetNodesWithOrder(store.CloneBudgetTree(tree))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_nodes (
				id, company_id, name, parent_id, budget, consumed, reserved_pool, period, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				reserved_pool = EXCLUDED.reserved_pool,
				period = EXCLUDED.period,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, row.ID, companyID, row.Name, row.ParentID, row.Budget, row.Consumed,
			row.ReservedPool, row.Period, row.sortOrder); err != nil {
			return fmt.Errorf("upsert budget node %s: %w", row.ID, err)
		}
	}
	return pruneByIDForCompany(ctx, r.db, "budget_nodes", companyID, ids)
}
