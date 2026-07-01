package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) Groups(ctx context.Context) ([]types.BudgetGroup, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, budget, consumed FROM budget_groups WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.BudgetGroup, 0)
	for rows.Next() {
		var group types.BudgetGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Budget, &group.Consumed); err != nil {
			return nil, err
		}
		memberRows, err := r.db.Query(ctx, `
			SELECT member_id FROM budget_group_members WHERE company_id = $1 AND group_id = $2 ORDER BY member_id
		`, companyID, group.ID)
		if err == nil {
			for memberRows.Next() {
				var memberID string
				if err := memberRows.Scan(&memberID); err == nil {
					group.MemberIDs = append(group.MemberIDs, memberID)
				}
			}
			memberRows.Close()
		}
		deptRows, err := r.db.Query(ctx, `
			SELECT department_id FROM budget_group_departments WHERE company_id = $1 AND group_id = $2 ORDER BY department_id
		`, companyID, group.ID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneBudgetGroups(items), nil
}

func (r *pgBudgetRepo) SetGroups(ctx context.Context, groups []types.BudgetGroup) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneBudgetGroups(groups)
	ids := make([]string, len(cloned))
	for i, group := range cloned {
		ids[i] = group.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_groups (id, company_id, name, budget, consumed, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				updated_at = NOW()
		`, group.ID, companyID, group.Name, group.Budget, group.Consumed); err != nil {
			return fmt.Errorf("upsert budget group %s: %w", group.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_members WHERE company_id = $1 AND group_id = $2`, companyID, group.ID); err != nil {
			return err
		}
		for _, memberID := range group.MemberIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO budget_group_members (company_id, group_id, member_id) VALUES ($1, $2, $3)
			`, companyID, group.ID, memberID); err != nil {
				return err
			}
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_departments WHERE company_id = $1 AND group_id = $2`, companyID, group.ID); err != nil {
			return err
		}
		for _, deptID := range group.DepartmentIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO budget_group_departments (company_id, group_id, department_id) VALUES ($1, $2, $3)
			`, companyID, group.ID, deptID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_members WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_departments WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM budget_groups WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "budget_group_members", "group_id", companyID, ids); err != nil {
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "budget_group_departments", "group_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "budget_groups", companyID, ids)
}
