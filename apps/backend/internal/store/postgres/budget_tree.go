package postgres

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) Tree() []types.BudgetNode {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, parent_id, budget, consumed, reserved_pool, period, sort_order
		FROM budget_nodes ORDER BY sort_order
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	flat := make([]flatBudgetNode, 0)
	for rows.Next() {
		var row flatBudgetNode
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.Budget, &row.Consumed,
			&row.ReservedPool, &row.Period, &row.sortOrder,
		); err != nil {
			return nil
		}
		flat = append(flat, row)
	}
	return store.CloneBudgetTree(buildBudgetTree(flat))
}

func (r *pgBudgetRepo) SetTree(tree []types.BudgetNode) error {
	flat := flattenBudgetNodesWithOrder(store.CloneBudgetTree(tree))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO budget_nodes (
				id, name, parent_id, budget, consumed, reserved_pool, period, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				reserved_pool = EXCLUDED.reserved_pool,
				period = EXCLUDED.period,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, row.ID, row.Name, row.ParentID, row.Budget, row.Consumed,
			row.ReservedPool, row.Period, row.sortOrder); err != nil {
			return fmt.Errorf("upsert budget node %s: %w", row.ID, err)
		}
	}
	return pruneByID(r.ctx, r.db, "budget_nodes", ids)
}
