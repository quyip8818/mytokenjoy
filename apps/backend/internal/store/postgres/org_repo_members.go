package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id, personal_quota
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
			&item.PersonalQuota,
		); err != nil {
			return nil, err
		}
		item.CompanyID = companyID
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
	return items, nil
}

func (r *pgOrgRepo) MemberByID(ctx context.Context, memberID string) (*types.Member, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id, personal_quota
		FROM members WHERE company_id = $1 AND id = $2
	`, companyID, memberID)
	var item types.Member
	if err := row.Scan(
		&item.ID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalQuota,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CompanyID = companyID
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
	return &item, nil
}

func (r *pgOrgRepo) MemberByEmail(ctx context.Context, companyID int64, email string) (*types.Member, string, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id, personal_quota, password_hash
		FROM members WHERE company_id = $1 AND email = $2
	`, companyID, email)
	var item types.Member
	var passwordHash *string
	if err := row.Scan(
		&item.ID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalQuota, &passwordHash,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", err
	}
	item.CompanyID = companyID
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
	hash := ""
	if passwordHash != nil {
		hash = *passwordHash
	}
	return &item, hash, nil
}

func (r *pgOrgRepo) GetMemberAuthz(ctx context.Context, companyID int64, memberID string) (*store.MemberAuthz, error) {
	row := r.db.QueryRow(ctx, `
		SELECT m.id, m.name, m.phone, m.email, m.department_id, m.department_name, m.status, m.source, m.external_id, m.personal_quota,
		       c.authz_revision
		FROM members m
		JOIN companies c ON c.id = m.company_id
		WHERE m.company_id = $1 AND m.id = $2
	`, companyID, memberID)
	var item types.Member
	var revision int64
	if err := row.Scan(
		&item.ID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalQuota, &revision,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CompanyID = companyID
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
	roles, err := r.rolesForCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &store.MemberAuthz{
		Member:        item,
		Roles:         roles,
		AuthzRevision: revision,
	}, nil
}

func (r *pgOrgRepo) MemberPersonalQuota(ctx context.Context, memberID string) (float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var quota float64
	err := r.db.QueryRow(ctx, `
		SELECT personal_quota FROM members WHERE company_id = $1 AND id = $2
	`, companyID, memberID).Scan(&quota)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return quota, true, nil
}

func (r *pgOrgRepo) SetMembers(ctx context.Context, members []types.Member) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneMembers(members)
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
				status, source, external_id, personal_quota, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				email = EXCLUDED.email,
				department_id = EXCLUDED.department_id,
				department_name = EXCLUDED.department_name,
				status = EXCLUDED.status,
				source = EXCLUDED.source,
				external_id = EXCLUDED.external_id,
				personal_quota = EXCLUDED.personal_quota,
				updated_at = NOW()
		`, member.ID, companyID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID, member.PersonalQuota); err != nil {
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

func (r *pgOrgRepo) UpdateMemberPersonalQuota(ctx context.Context, memberID string, personalQuota float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE members SET personal_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, memberID, personalQuota)
	if err != nil {
		return fmt.Errorf("update member personal quota: %w", err)
	}
	return nil
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
