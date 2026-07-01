package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Departments(ctx context.Context) ([]types.Department, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, parent_id, member_count, external_id, source, manager_id, sort_order
		FROM departments
		WHERE company_id = $1
		ORDER BY sort_order
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flat := make([]flatDepartment, 0)
	for rows.Next() {
		var row flatDepartment
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.MemberCount,
			&row.ExternalID, &row.Source, &row.ManagerID, &row.sortOrder,
		); err != nil {
			return nil, err
		}
		flat = append(flat, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneDepartments(buildDepartmentTree(flat)), nil
}

func (r *pgOrgRepo) SetDepartments(ctx context.Context, departments []types.Department) error {
	companyID := store.CompanyID(ctx)
	flat := flattenDepartmentsWithOrder(store.CloneDepartments(departments))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO departments (
				id, company_id, name, parent_id, member_count, external_id, source, manager_id, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				member_count = EXCLUDED.member_count,
				external_id = EXCLUDED.external_id,
				source = EXCLUDED.source,
				manager_id = EXCLUDED.manager_id,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, row.ID, companyID, row.Name, row.ParentID, row.MemberCount,
			row.ExternalID, row.Source, row.ManagerID, row.sortOrder); err != nil {
			return fmt.Errorf("upsert department %s: %w", row.ID, err)
		}
	}
	return pruneByIDForCompany(ctx, r.db, "departments", companyID, ids)
}
