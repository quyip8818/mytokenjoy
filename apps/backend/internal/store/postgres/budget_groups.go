package postgres

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) Groups() []types.BudgetGroup {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, budget, consumed FROM budget_groups ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.BudgetGroup, 0)
	for rows.Next() {
		var group types.BudgetGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Budget, &group.Consumed); err != nil {
			return nil
		}
		memberRows, err := r.db.Query(r.ctx, `
			SELECT member_id FROM budget_group_members WHERE group_id = $1 ORDER BY member_id
		`, group.ID)
		if err == nil {
			for memberRows.Next() {
				var memberID string
				if err := memberRows.Scan(&memberID); err == nil {
					group.MemberIDs = append(group.MemberIDs, memberID)
				}
			}
			memberRows.Close()
		}
		deptRows, err := r.db.Query(r.ctx, `
			SELECT department_id FROM budget_group_departments WHERE group_id = $1 ORDER BY department_id
		`, group.ID)
		if err == nil {
			for deptRows.Next() {
				var deptID string
				if err := deptRows.Scan(&deptID); err == nil {
					group.DepartmentIDs = append(group.DepartmentIDs, deptID)
				}
			}
			deptRows.Close()
		}
		items = append(items, group)
	}
	return store.CloneBudgetGroups(items)
}

func (r *pgBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	cloned := store.CloneBudgetGroups(groups)
	ids := make([]string, len(cloned))
	for i, group := range cloned {
		ids[i] = group.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO budget_groups (id, name, budget, consumed, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				updated_at = NOW()
		`, group.ID, group.Name, group.Budget, group.Consumed); err != nil {
			return fmt.Errorf("upsert budget group %s: %w", group.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM budget_group_members WHERE group_id = $1`, group.ID); err != nil {
			return err
		}
		for _, memberID := range group.MemberIDs {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO budget_group_members (group_id, member_id) VALUES ($1, $2)
			`, group.ID, memberID); err != nil {
				return err
			}
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM budget_group_departments WHERE group_id = $1`, group.ID); err != nil {
			return err
		}
		for _, deptID := range group.DepartmentIDs {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO budget_group_departments (group_id, department_id) VALUES ($1, $2)
			`, group.ID, deptID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM budget_group_members`); err != nil {
			return err
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM budget_group_departments`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM budget_groups`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "budget_group_members", "group_id", ids); err != nil {
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "budget_group_departments", "group_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "budget_groups", ids)
}
