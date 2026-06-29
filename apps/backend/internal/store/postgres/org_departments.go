package postgres

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Departments() []types.Department {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, parent_id, member_count, external_id, source, manager_id, sort_order
		FROM departments
		ORDER BY sort_order
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	flat := make([]flatDepartment, 0)
	for rows.Next() {
		var row flatDepartment
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.MemberCount,
			&row.ExternalID, &row.Source, &row.ManagerID, &row.sortOrder,
		); err != nil {
			return nil
		}
		flat = append(flat, row)
	}
	return store.CloneDepartments(buildDepartmentTree(flat))
}

func (r *pgOrgRepo) SetDepartments(departments []types.Department) error {
	flat := flattenDepartmentsWithOrder(store.CloneDepartments(departments))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO departments (
				id, name, parent_id, member_count, external_id, source, manager_id, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				member_count = EXCLUDED.member_count,
				external_id = EXCLUDED.external_id,
				source = EXCLUDED.source,
				manager_id = EXCLUDED.manager_id,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, row.ID, row.Name, row.ParentID, row.MemberCount,
			row.ExternalID, row.Source, row.ManagerID, row.sortOrder); err != nil {
			return fmt.Errorf("upsert department %s: %w", row.ID, err)
		}
	}
	return pruneByID(r.ctx, r.db, "departments", ids)
}
