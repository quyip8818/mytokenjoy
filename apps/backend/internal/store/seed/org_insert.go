package seed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
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

func insertOrgNodes(ctx context.Context, exec tableWriter, tid int64, nodes []types.OrgNode) error {
	flat := pkgorg.FlattenOrgNodeTree(nodes)
	for i, node := range flat {
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_nodes (
				id, company_id, name, parent_id, member_count, external_id, source, manager_id, sort_order,
				budget, consumed, reserved_pool, period, default_model, fallback_model, routing_inherited
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (company_id, id) DO NOTHING
		`, node.ID, tid, node.Name, node.ParentID, node.MemberCount,
			node.ExternalID, node.Source, node.ManagerID, i,
			node.Budget, node.Consumed, node.ReservedPool, node.Period,
			node.DefaultModel, node.FallbackModel, node.RoutingInherited); err != nil {
			return fmt.Errorf("seed org node %s: %w", node.ID, err)
		}
	}
	return nil
}

func insertModelAllowlist(ctx context.Context, exec tableWriter, tid int64, rows []store.ModelAllowlistRow) error {
	for _, row := range rows {
		if _, err := exec.Exec(ctx, `
			INSERT INTO model_allowlist (company_id, owner_type, owner_id, model_name)
			VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING
		`, tid, row.OwnerType, row.OwnerID, row.ModelName); err != nil {
			return fmt.Errorf("seed allowlist %s/%s: %w", row.OwnerType, row.OwnerID, err)
		}
	}
	return nil
}

func insertMembers(ctx context.Context, exec tableWriter, tid int64, members []types.Member, roleIDByName map[string]string) error {
	demoHash := DemoPasswordHash()
	for _, member := range members {
		var passwordHash *string
		if member.Status == "active" && member.Email != "" {
			hash := demoHash
			passwordHash = &hash
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO members (
				id, company_id, name, phone, email, department_id, department_name,
				status, source, external_id, personal_quota, password_hash
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (company_id, id) DO NOTHING
		`, member.ID, member.CompanyID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID, member.PersonalQuota, passwordHash); err != nil {
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

func insertOrgIntegration(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
	integration := snap.OrgIntegration
	var platform *string
	if integration.Platform != nil {
		s := string(*integration.Platform)
		platform = &s
	}
	fieldMappingsJSON, err := json.Marshal(integration.FieldMappings)
	if err != nil {
		return fmt.Errorf("marshal field mappings: %w", err)
	}
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_integration (
			company_id, platform, connected,
			enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im,
			encrypted_credential, field_mappings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (company_id) DO NOTHING
	`, tid, platform, integration.Connected,
		integration.Enabled, integration.StartTime, integration.FrequencyHours,
		integration.DeleteMemberThreshold, integration.DeleteDepartmentThreshold,
		integration.NotifyPhone, integration.NotifyEmail, integration.NotifyIm,
		integration.EncryptedCredential, fieldMappingsJSON); err != nil {
		return err
	}
	for _, log := range snap.SyncLogs {
		t, err := pkgtime.Parse(log.Time)
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
