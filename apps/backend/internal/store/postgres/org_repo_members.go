package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

const memberSelect = `
	SELECT m.id, m.user_id, m.name, COALESCE(u.phone,''), COALESCE(u.email,''), COALESCE(m.department_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(o.name, ''), m.status, m.source, m.external_id, m.personal_budget
	FROM members m
	JOIN users u ON u.id = m.user_id
	LEFT JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
`

const memberListSelect = `
	SELECT m.id, m.user_id, m.name, COALESCE(u.phone,''), COALESCE(u.email,''), COALESCE(m.department_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(o.name, ''),
		m.status, m.source, m.external_id, m.personal_budget,
		COALESCE(array_agg(r.name ORDER BY r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
	FROM members m
	JOIN users u ON u.id = m.user_id
	LEFT JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
	LEFT JOIN member_roles mr ON mr.company_id = m.company_id AND mr.member_id = m.id
	LEFT JOIN roles r ON r.id = mr.role_id
`

func (r *pgOrgRepo) FindMemberCompanyID(ctx context.Context, memberID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT company_id FROM members WHERE id = $1 LIMIT 1`, memberID).Scan(&companyID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, nil
	}
	return companyID, err
}

func (r *pgOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, memberListSelect+`
		WHERE m.company_id = $1
		GROUP BY m.id, m.user_id, m.name, u.phone, u.email, m.department_id, o.name,
			m.status, m.source, m.external_id, m.personal_budget
		ORDER BY m.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Member, 0)
	for rows.Next() {
		var item types.Member
		var roles []string
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Name, &item.Phone, &item.Email,
			&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
			&item.PersonalBudget, &roles,
		); err != nil {
			return nil, err
		}
		item.CompanyID = companyID
		item.Roles = roles
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgOrgRepo) MemberByID(ctx context.Context, memberID uuid.UUID) (*types.Member, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, memberSelect+` WHERE m.company_id = $1 AND m.id = $2`, companyID, memberID)
	var item types.Member
	if err := row.Scan(
		&item.ID, &item.UserID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalBudget,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CompanyID = companyID
	roleRows, err := r.db.Query(ctx, `
		SELECT ro.name FROM member_roles mr
		JOIN roles ro ON ro.id = mr.role_id
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

func (r *pgOrgRepo) MemberByEmail(ctx context.Context, companyID uuid.UUID, email string) (*types.Member, string, error) {
	row := r.db.QueryRow(ctx, `
		SELECT m.id, m.user_id, m.name, COALESCE(u.phone,''), COALESCE(u.email,''), COALESCE(m.department_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(o.name, ''), m.status, m.source, m.external_id, m.personal_budget, u.password_hash
		FROM members m
		JOIN users u ON u.id = m.user_id
		LEFT JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
		WHERE m.company_id = $1 AND u.email = $2`, companyID, email)
	var item types.Member
	var passwordHash *string
	if err := row.Scan(
		&item.ID, &item.UserID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalBudget, &passwordHash,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", err
	}
	item.CompanyID = companyID
	roleRows, err := r.db.Query(ctx, `
		SELECT ro.name FROM member_roles mr
		JOIN roles ro ON ro.id = mr.role_id
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

func (r *pgOrgRepo) GetMemberAuthz(ctx context.Context, companyID uuid.UUID, memberID uuid.UUID) (*store.MemberAuthz, error) {
	row := r.db.QueryRow(ctx, `
		SELECT m.id, m.user_id, m.name, COALESCE(u.phone,''), COALESCE(u.email,''), COALESCE(m.department_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(o.name, ''), m.status, m.source, m.external_id, m.personal_budget,
		       c.authz_revision
		FROM members m
		JOIN users u ON u.id = m.user_id
		JOIN companies c ON c.id = m.company_id
		LEFT JOIN org_nodes o ON o.company_id = m.company_id AND o.id = m.department_id
		WHERE m.company_id = $1 AND m.id = $2
	`, companyID, memberID)
	var item types.Member
	var revision int64
	if err := row.Scan(
		&item.ID, &item.UserID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalBudget, &revision,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CompanyID = companyID
	roleRows, err := r.db.Query(ctx, `
		SELECT ro.name FROM member_roles mr
		JOIN roles ro ON ro.id = mr.role_id
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

func (r *pgOrgRepo) MemberPersonalBudget(ctx context.Context, memberID uuid.UUID) (int64, bool, error) {
	companyID := store.CompanyID(ctx)
	var quota int64
	err := r.db.QueryRow(ctx, `
		SELECT personal_budget FROM members WHERE company_id = $1 AND id = $2
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
	ids := make([]uuid.UUID, len(cloned))
	for i, member := range cloned {
		ids[i] = member.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO members (
				id, company_id, user_id, name, department_id,
				status, source, external_id, personal_budget, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				user_id = EXCLUDED.user_id,
				department_id = EXCLUDED.department_id,
				status = EXCLUDED.status,
				source = EXCLUDED.source,
				external_id = EXCLUDED.external_id,
				personal_budget = EXCLUDED.personal_budget,
				updated_at = NOW()
		`, member.ID, companyID, member.UserID, member.Name,
			nilUUID(member.DepartmentID), member.Status, member.Source, member.ExternalID, member.PersonalBudget); err != nil {
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
	if _, err := r.db.Exec(ctx, `
		UPDATE platform_keys SET member_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND member_id IS NOT NULL AND NOT (member_id = ANY($2))
	`, companyID, ids); err != nil {
		return fmt.Errorf("detach platform keys from pruned members: %w", err)
	}
	return pruneByIDForCompany(ctx, r.db, "members", companyID, ids)
}

func (r *pgOrgRepo) UpdateMemberPersonalBudget(ctx context.Context, memberID uuid.UUID, personalBudget int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE members SET personal_budget = $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, memberID, personalBudget)
	if err != nil {
		return fmt.Errorf("update member personal budget: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $3, updated_at = NOW()
		WHERE id = (SELECT user_id FROM members WHERE company_id = $1 AND id = $2)
	`, companyID, memberID, passwordHash)
	if err != nil {
		return fmt.Errorf("set user password: %w", err)
	}
	return nil
}
