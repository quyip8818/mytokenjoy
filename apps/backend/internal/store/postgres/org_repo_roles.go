package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Permissions(ctx context.Context) ([]types.Permission, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, grp FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Permission, 0)
	for rows.Next() {
		var item types.Permission
		if err := rows.Scan(&item.ID, &item.Name, &item.Group); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgOrgRepo) rolesForCompany(ctx context.Context, companyID uuid.UUID) ([]types.Role, error) {
	return r.rolesByCompanyID(ctx, companyID)
}

func loadRoleNameIndex(ctx context.Context, db dbQuerier, companyID uuid.UUID) (map[string]string, error) {
	rows, err := db.Query(ctx, `SELECT id, name FROM roles WHERE company_id = $1`, companyID)
	if err != nil {
		return nil, fmt.Errorf("load roles index: %w", err)
	}
	defer rows.Close()
	index := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		index[name] = id
	}
	return index, rows.Err()
}

func (r *pgOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
	return r.rolesByCompanyID(ctx, store.CompanyID(ctx))
}

func (r *pgOrgRepo) rolesByCompanyID(ctx context.Context, companyID uuid.UUID) ([]types.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.type, COUNT(mr.member_id)::int
		FROM roles r
		LEFT JOIN member_roles mr ON mr.company_id = r.company_id AND mr.role_id = r.id
		WHERE r.company_id = $1
		GROUP BY r.id, r.name, r.type
		ORDER BY r.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	type roleRow struct {
		role types.Role
	}
	batch := make([]roleRow, 0)
	for rows.Next() {
		var row roleRow
		if err := rows.Scan(&row.role.ID, &row.role.Name, &row.role.Type, &row.role.MemberCount); err != nil {
			rows.Close()
			return nil, err
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	items := make([]types.Role, 0, len(batch))
	for _, row := range batch {
		role := row.role
		grantRows, err := r.db.Query(ctx, `
			SELECT permission_id FROM role_permission_grants WHERE company_id = $1 AND role_id = $2 ORDER BY permission_id
		`, companyID, role.ID)
		if err != nil {
			return nil, err
		}
		for grantRows.Next() {
			var ref string
			if err := grantRows.Scan(&ref); err != nil {
				grantRows.Close()
				return nil, err
			}
			role.Permissions = append(role.Permissions, ref)
		}
		grantRows.Close()
		items = append(items, role)
	}
	return items, nil
}

func (r *pgOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneRoles(roles)
	ids := make([]uuid.UUID, len(cloned))
	for i, role := range cloned {
		ids[i] = role.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type) VALUES ($1, $2, $3, $4)
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type
		`, role.ID, companyID, role.Name, role.Type); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM role_permission_grants WHERE company_id = $1 AND role_id = $2`, companyID, role.ID); err != nil {
			return fmt.Errorf("clear role grants: %w", err)
		}
		for _, perm := range role.Permissions {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO role_permission_grants (company_id, role_id, permission_id) VALUES ($1, $2, $3)
			`, companyID, role.ID, perm); err != nil {
				return fmt.Errorf("insert role grant: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM role_permission_grants WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM roles WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "role_permission_grants", "role_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "roles", companyID, ids)
}
