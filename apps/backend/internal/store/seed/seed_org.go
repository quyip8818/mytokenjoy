package seed

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/timeparse"
)

func insertPermissions(ctx context.Context, exec tableWriter, permissions []types.Permission) error {
	for _, perm := range permissions {
		if _, err := exec.Exec(ctx, `
			INSERT INTO permissions (id, name, grp) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, perm.ID, perm.Name, perm.Group); err != nil {
			return fmt.Errorf("seed permission %s: %w", perm.ID, err)
		}
	}
	return nil
}

func insertRoles(ctx context.Context, exec tableWriter, tid int64, roles []types.Role) error {
	for _, role := range roles {
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, member_count) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (company_id, id) DO NOTHING
		`, role.ID, tid, role.Name, role.Type, role.MemberCount); err != nil {
			return fmt.Errorf("seed role %s: %w", role.ID, err)
		}
		for _, perm := range role.Permissions {
			if _, err := exec.Exec(ctx, `
				INSERT INTO role_permission_grants (company_id, role_id, permission_ref) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, role.ID, perm); err != nil {
				return fmt.Errorf("seed role grant: %w", err)
			}
		}
	}
	return nil
}

func buildRoleNameIndex(roles []types.Role) map[string]string {
	index := make(map[string]string, len(roles))
	for _, role := range roles {
		index[role.Name] = role.ID
	}
	return index
}

func insertDepartments(ctx context.Context, exec tableWriter, tid int64, departments []types.Department) error {
	flat := pkgorg.FlattenDepartmentTree(departments)
	for i, dept := range flat {
		if _, err := exec.Exec(ctx, `
			INSERT INTO departments (
				id, company_id, name, parent_id, member_count, external_id, source, manager_id, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (company_id, id) DO NOTHING
		`, dept.ID, tid, dept.Name, dept.ParentID, dept.MemberCount,
			dept.ExternalID, dept.Source, dept.ManagerID, i); err != nil {
			return fmt.Errorf("seed department %s: %w", dept.ID, err)
		}
	}
	return nil
}

func insertMembers(ctx context.Context, exec tableWriter, tid int64, members []types.Member, roleIDByName map[string]string) error {
	for _, member := range members {
		if _, err := exec.Exec(ctx, `
			INSERT INTO members (
				id, company_id, name, phone, email, department_id, department_name, status, source, external_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (company_id, id) DO NOTHING
		`, member.ID, member.CompanyID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID); err != nil {
			return fmt.Errorf("seed member %s: %w", member.ID, err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := exec.Exec(ctx, `
				INSERT INTO member_roles (company_id, member_id, role_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, member.CompanyID, member.ID, roleID); err != nil {
				return fmt.Errorf("seed member role: %w", err)
			}
		}
	}
	return nil
}

func insertOrgConfig(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
	var platform *string
	if snap.DataSourceStatus.Platform != nil {
		s := string(*snap.DataSourceStatus.Platform)
		platform = &s
	}
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_data_source_status (company_id, platform, connected)
		VALUES ($1, $2, $3) ON CONFLICT (company_id) DO NOTHING
	`, tid, platform, snap.DataSourceStatus.Connected); err != nil {
		return err
	}
	cfg := snap.SyncConfig
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_sync_config (
			company_id, enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (company_id) DO NOTHING
	`, tid, cfg.Enabled, cfg.StartTime, cfg.FrequencyHours,
		cfg.DeleteMemberThreshold, cfg.DeleteDepartmentThreshold,
		cfg.NotifyPhone, cfg.NotifyEmail, cfg.NotifyIm); err != nil {
		return err
	}
	for _, log := range snap.SyncLogs {
		t, err := timeparse.Parse(log.Time)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_sync_logs (id, company_id, time, type, result, detail)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (company_id, id) DO NOTHING
		`, log.ID, tid, t, log.Type, log.Result, log.Detail); err != nil {
			return err
		}
	}
	for _, failure := range snap.ImportFailures {
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_import_failures (id, company_id, name, employee_id, reason)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (company_id, id) DO NOTHING
		`, failure.ID, tid, failure.Name, failure.EmployeeID, failure.Reason); err != nil {
			return err
		}
	}
	return nil
}
