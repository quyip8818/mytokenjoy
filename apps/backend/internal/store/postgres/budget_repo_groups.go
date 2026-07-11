package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) GetGroupBudget(ctx context.Context, groupID string) (float64, float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var budget float64
	err := r.db.QueryRow(ctx, `
		SELECT budget FROM budget_groups WHERE company_id = $1 AND id = $2
	`, companyID, groupID).Scan(&budget)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return budget, 0, true, nil
}

func (r *pgBudgetRepo) Groups(ctx context.Context) ([]types.BudgetGroup, error) {
	companyID := store.CompanyID(ctx)

	rows, err := r.db.Query(ctx, `SELECT id, name, budget FROM budget_groups WHERE company_id = $1 ORDER BY id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]types.BudgetGroup, 0)
	groupIndex := make(map[string]int)
	for rows.Next() {
		var g types.BudgetGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Budget); err != nil {
			return nil, err
		}
		g.Consumed = 0
		g.MemberIDs = []string{}
		g.DepartmentIDs = []string{}
		groupIndex[g.ID] = len(groups)
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}

	memberRows, err := r.db.Query(ctx, `SELECT group_id, member_id FROM budget_group_members WHERE company_id = $1 ORDER BY group_id, member_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()
	for memberRows.Next() {
		var groupID, memberID string
		if err := memberRows.Scan(&groupID, &memberID); err != nil {
			return nil, err
		}
		if idx, ok := groupIndex[groupID]; ok {
			groups[idx].MemberIDs = append(groups[idx].MemberIDs, memberID)
		}
	}
	if err := memberRows.Err(); err != nil {
		return nil, err
	}

	deptRows, err := r.db.Query(ctx, `SELECT group_id, department_id FROM budget_group_departments WHERE company_id = $1 ORDER BY group_id, department_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var groupID, deptID string
		if err := deptRows.Scan(&groupID, &deptID); err != nil {
			return nil, err
		}
		if idx, ok := groupIndex[groupID]; ok {
			groups[idx].DepartmentIDs = append(groups[idx].DepartmentIDs, deptID)
		}
	}
	if err := deptRows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *pgBudgetRepo) SetGroups(ctx context.Context, groups []types.BudgetGroup) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneBudgetGroups(groups)
	ids := make([]string, len(cloned))
	for i, group := range cloned {
		ids[i] = group.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_groups (id, company_id, name, budget, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				updated_at = NOW()
		`, group.ID, companyID, group.Name, group.Budget); err != nil {
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
	if _, err := r.db.Exec(ctx, `
		UPDATE platform_keys SET budget_group_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND budget_group_id IS NOT NULL AND NOT (budget_group_id = ANY($2))
	`, companyID, ids); err != nil {
		return fmt.Errorf("detach platform keys from pruned budget groups: %w", err)
	}
	return pruneByIDForCompany(ctx, r.db, "budget_groups", companyID, ids)
}
