package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `SELECT id, name, type, member_count FROM roles WHERE company_id = $1 ORDER BY id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Role, 0)
	for rows.Next() {
		var role types.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Type, &role.MemberCount); err != nil {
			return nil, err
		}
		grantRows, err := r.db.Query(ctx, `
			SELECT permission_ref FROM role_permission_grants WHERE company_id = $1 AND role_id = $2 ORDER BY permission_ref
		`, companyID, role.ID)
		if err == nil {
			for grantRows.Next() {
				var ref string
				if err := grantRows.Scan(&ref); err == nil {
					role.Permissions = append(role.Permissions, ref)
				}
			}
			grantRows.Close()
		}
		items = append(items, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneRoles(items), nil
}

func (r *pgOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneRoles(roles)
	ids := make([]string, len(cloned))
	for i, role := range cloned {
		ids[i] = role.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, member_count) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				member_count = EXCLUDED.member_count
		`, role.ID, companyID, role.Name, role.Type, role.MemberCount); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM role_permission_grants WHERE company_id = $1 AND role_id = $2`, companyID, role.ID); err != nil {
			return fmt.Errorf("clear role grants: %w", err)
		}
		for _, perm := range role.Permissions {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO role_permission_grants (company_id, role_id, permission_ref) VALUES ($1, $2, $3)
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
