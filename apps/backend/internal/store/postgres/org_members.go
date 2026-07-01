package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id
		FROM members WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Member, 0)
	for rows.Next() {
		var item types.Member
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Phone, &item.Email,
			&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		); err != nil {
			return nil, err
		}
		roleRows, err := r.db.Query(ctx, `
			SELECT ro.name FROM member_roles mr
			JOIN roles ro ON ro.company_id = mr.company_id AND ro.id = mr.role_id
			WHERE mr.company_id = $1 AND mr.member_id = $2
			ORDER BY ro.name
		`, companyID, item.ID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneMembers(items), nil
}

func (r *pgOrgRepo) SetMembers(ctx context.Context, members []types.Member) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneMembers(members)
	roleIDByName, err := loadRoleNameIndex(ctx, r.db, companyID)
	if err != nil {
		return err
	}
	ids := make([]string, len(cloned))
	for i, member := range cloned {
		ids[i] = member.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO members (
				id, company_id, name, phone, email, department_id, department_name,
				status, source, external_id, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				email = EXCLUDED.email,
				department_id = EXCLUDED.department_id,
				department_name = EXCLUDED.department_name,
				status = EXCLUDED.status,
				source = EXCLUDED.source,
				external_id = EXCLUDED.external_id,
				updated_at = NOW()
		`, member.ID, companyID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID); err != nil {
			return fmt.Errorf("upsert member %s: %w", member.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM member_roles WHERE company_id = $1 AND member_id = $2`, companyID, member.ID); err != nil {
			return fmt.Errorf("clear member roles: %w", err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := r.db.Exec(ctx, `
				INSERT INTO member_roles (company_id, member_id, role_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, companyID, member.ID, roleID); err != nil {
				return fmt.Errorf("insert member role: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM member_roles WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM members WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "member_roles", "member_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "members", companyID, ids)
}

func loadRoleNameIndex(ctx context.Context, db dbQuerier, companyID int64) (map[string]string, error) {
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

func (r *pgOrgRepo) SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE members SET password_hash = $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, memberID, passwordHash)
	if err != nil {
		return fmt.Errorf("set member password: %w", err)
	}
	return nil
}
