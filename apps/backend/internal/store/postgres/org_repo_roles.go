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
	rows, err := db.Query(ctx, `
		SELECT id::text, name FROM roles WHERE company_id IS NULL OR company_id = $1
	`, companyID)
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
	// ponytail: OR 条件可能不走索引，当前角色数极少无影响；
	// 若将来角色数 >1000，改为 UNION ALL（全局预设 + 本公司自定义分开查）。
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.type, r.permissions, COUNT(mr.member_id)::int AS member_count
		FROM roles r
		LEFT JOIN member_roles mr ON mr.role_id = r.id AND mr.company_id = $1
		WHERE r.company_id IS NULL OR r.company_id = $1
		GROUP BY r.id, r.name, r.type, r.permissions
		ORDER BY r.type, r.name
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []types.Role
	for rows.Next() {
		var role types.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Type, &role.Permissions, &role.MemberCount); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *pgOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
	companyID := store.CompanyID(ctx)
	ids := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		if role.Type == "preset" {
			continue // 不操作全局预设行
		}
		ids = append(ids, role.ID)
		if _, err := r.db.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, permissions)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (company_id, name) DO UPDATE SET permissions = EXCLUDED.permissions
		`, role.ID, companyID, role.Name, role.Type, role.Permissions); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.ID, err)
		}
	}
	// Prune deleted custom roles.
	// 安全说明：pruneByIDForCompany 使用 DELETE FROM roles WHERE company_id = $1 AND id NOT IN (...)，
	// 全局预设角色 company_id IS NULL 不会被 company_id = $1 匹配到，不会被误删。
	if len(ids) == 0 {
		_, err := r.db.Exec(ctx, `DELETE FROM roles WHERE company_id = $1`, companyID)
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "roles", companyID, ids)
}
