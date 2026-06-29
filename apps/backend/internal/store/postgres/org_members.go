package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Members() []types.Member {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id
		FROM members ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.Member, 0)
	for rows.Next() {
		var item types.Member
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Phone, &item.Email,
			&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		); err != nil {
			return nil
		}
		roleRows, err := r.db.Query(r.ctx, `
			SELECT ro.name FROM member_roles mr
			JOIN roles ro ON ro.id = mr.role_id
			WHERE mr.member_id = $1
			ORDER BY ro.name
		`, item.ID)
		if err == nil {
			for roleRows.Next() {
				var name string
				if err := roleRows.Scan(&name); err == nil {
					item.Roles = append(item.Roles, name)
				}
			}
			roleRows.Close()
		}
		items = append(items, item)
	}
	return store.CloneMembers(items)
}

func (r *pgOrgRepo) SetMembers(members []types.Member) error {
	cloned := store.CloneMembers(members)
	roleIDByName, err := loadRoleNameIndex(r.ctx, r.db)
	if err != nil {
		return err
	}
	ids := make([]string, len(cloned))
	for i, member := range cloned {
		ids[i] = member.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO members (
				id, name, phone, email, department_id, department_name,
				status, source, external_id, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				email = EXCLUDED.email,
				department_id = EXCLUDED.department_id,
				department_name = EXCLUDED.department_name,
				status = EXCLUDED.status,
				source = EXCLUDED.source,
				external_id = EXCLUDED.external_id,
				updated_at = NOW()
		`, member.ID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID); err != nil {
			return fmt.Errorf("upsert member %s: %w", member.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM member_roles WHERE member_id = $1`, member.ID); err != nil {
			return fmt.Errorf("clear member roles: %w", err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO member_roles (member_id, role_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, member.ID, roleID); err != nil {
				return fmt.Errorf("insert member role: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM member_roles`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM members`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "member_roles", "member_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "members", ids)
}

func loadRoleNameIndex(ctx context.Context, db dbQuerier) (map[string]string, error) {
	rows, err := db.Query(ctx, `SELECT id, name FROM roles`)
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
