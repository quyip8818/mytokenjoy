package postgres

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Roles() []types.Role {
	rows, err := r.db.Query(r.ctx, `SELECT id, name, type, member_count FROM roles ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.Role, 0)
	for rows.Next() {
		var role types.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Type, &role.MemberCount); err != nil {
			return nil
		}
		grantRows, err := r.db.Query(r.ctx, `
			SELECT permission_ref FROM role_permission_grants WHERE role_id = $1 ORDER BY permission_ref
		`, role.ID)
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
	return store.CloneRoles(items)
}

func (r *pgOrgRepo) SetRoles(roles []types.Role) error {
	cloned := store.CloneRoles(roles)
	ids := make([]string, len(cloned))
	for i, role := range cloned {
		ids[i] = role.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO roles (id, name, type, member_count) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				member_count = EXCLUDED.member_count
		`, role.ID, role.Name, role.Type, role.MemberCount); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM role_permission_grants WHERE role_id = $1`, role.ID); err != nil {
			return fmt.Errorf("clear role grants: %w", err)
		}
		for _, perm := range role.Permissions {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO role_permission_grants (role_id, permission_ref) VALUES ($1, $2)
			`, role.ID, perm); err != nil {
				return fmt.Errorf("insert role grant: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM role_permission_grants`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM roles`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "role_permission_grants", "role_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "roles", ids)
}
