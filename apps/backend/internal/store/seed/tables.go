package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/timeparse"
	"github.com/tokenjoy/backend/internal/store/treeutil"
)

type tableWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func ApplyTables(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	if err := insertPermissions(ctx, exec, snap.Permissions); err != nil {
		return err
	}
	if err := insertRoles(ctx, exec, snap.Roles); err != nil {
		return err
	}
	roleIDByName := buildRoleNameIndex(snap.Roles)
	if err := insertDepartments(ctx, exec, snap.Departments); err != nil {
		return err
	}
	if err := insertMembers(ctx, exec, snap.Members, roleIDByName); err != nil {
		return err
	}
	if err := insertOrgConfig(ctx, exec, snap); err != nil {
		return err
	}
	if err := insertBudget(ctx, exec, snap); err != nil {
		return err
	}
	if err := insertModels(ctx, exec, snap.Models); err != nil {
		return err
	}
	if err := insertRoutingRules(ctx, exec, snap.RoutingRules); err != nil {
		return err
	}
	if err := insertKeys(ctx, exec, snap); err != nil {
		return err
	}
	if err := insertAudit(ctx, exec, snap); err != nil {
		return err
	}
	return nil
}

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

func insertRoles(ctx context.Context, exec tableWriter, roles []types.Role) error {
	for _, role := range roles {
		if _, err := exec.Exec(ctx, `
			INSERT INTO roles (id, name, type, member_count) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, role.ID, role.Name, role.Type, role.MemberCount); err != nil {
			return fmt.Errorf("seed role %s: %w", role.ID, err)
		}
		for _, perm := range role.Permissions {
			if _, err := exec.Exec(ctx, `
				INSERT INTO role_permission_grants (role_id, permission_ref) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, role.ID, perm); err != nil {
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

func insertDepartments(ctx context.Context, exec tableWriter, departments []types.Department) error {
	flat := pkgorg.FlattenDepartmentTree(departments)
	for i, dept := range flat {
		if _, err := exec.Exec(ctx, `
			INSERT INTO departments (
				id, name, parent_id, member_count, external_id, source, manager_id, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, dept.ID, dept.Name, dept.ParentID, dept.MemberCount,
			dept.ExternalID, dept.Source, dept.ManagerID, i); err != nil {
			return fmt.Errorf("seed department %s: %w", dept.ID, err)
		}
	}
	return nil
}

func insertMembers(ctx context.Context, exec tableWriter, members []types.Member, roleIDByName map[string]string) error {
	for _, member := range members {
		if _, err := exec.Exec(ctx, `
			INSERT INTO members (
				id, name, phone, email, department_id, department_name, status, source, external_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, member.ID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID); err != nil {
			return fmt.Errorf("seed member %s: %w", member.ID, err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := exec.Exec(ctx, `
				INSERT INTO member_roles (member_id, role_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, member.ID, roleID); err != nil {
				return fmt.Errorf("seed member role: %w", err)
			}
		}
	}
	return nil
}

func insertOrgConfig(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	var platform *string
	if snap.DataSourceStatus.Platform != nil {
		s := string(*snap.DataSourceStatus.Platform)
		platform = &s
	}
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_data_source_status (id, platform, connected)
		VALUES (1, $1, $2) ON CONFLICT (id) DO NOTHING
	`, platform, snap.DataSourceStatus.Connected); err != nil {
		return err
	}
	cfg := snap.SyncConfig
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_sync_config (
			id, enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im
		) VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`, cfg.Enabled, cfg.StartTime, cfg.FrequencyHours,
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
			INSERT INTO org_sync_logs (id, time, type, result, detail)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING
		`, log.ID, t, log.Type, log.Result, log.Detail); err != nil {
			return err
		}
	}
	for _, failure := range snap.ImportFailures {
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_import_failures (id, name, employee_id, reason)
			VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING
		`, failure.ID, failure.Name, failure.EmployeeID, failure.Reason); err != nil {
			return err
		}
	}
	return nil
}

func insertBudget(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	flat := treeutil.FlattenBudgetTree(snap.BudgetTree)
	for i, node := range flat {
		if _, err := exec.Exec(ctx, `
			INSERT INTO budget_nodes (
				id, name, parent_id, budget, consumed, reserved_pool, period, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, node.ID, node.Name, node.ParentID, node.Budget, node.Consumed, node.ReservedPool, node.Period, i); err != nil {
			return err
		}
	}
	for _, group := range snap.BudgetGroups {
		if _, err := exec.Exec(ctx, `
			INSERT INTO budget_groups (id, name, budget, consumed)
			VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING
		`, group.ID, group.Name, group.Budget, group.Consumed); err != nil {
			return err
		}
		for _, memberID := range group.MemberIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO budget_group_members (group_id, member_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, group.ID, memberID); err != nil {
				return err
			}
		}
		for _, deptID := range group.DepartmentIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO budget_group_departments (group_id, department_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, group.ID, deptID); err != nil {
				return err
			}
		}
	}
	policy := snap.OverrunPolicy
	if _, err := exec.Exec(ctx, `
		INSERT INTO overrun_policy (id, thresholds, notify_email, notify_phone, notify_im, block_message)
		VALUES (1, $1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING
	`, policy.Thresholds, policy.NotifyEmail, policy.NotifyPhone, policy.NotifyIm, policy.BlockMessage); err != nil {
		return err
	}
	for _, rule := range snap.AlertRules {
		if _, err := exec.Exec(ctx, `
			INSERT INTO alert_rules (id, node_id, node_name, thresholds, enabled)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING
		`, rule.ID, rule.NodeID, rule.NodeName, rule.Thresholds, rule.Enabled); err != nil {
			return err
		}
		for _, roleID := range rule.NotifyRoleIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO alert_rule_notify_roles (rule_id, role_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, rule.ID, roleID); err != nil {
				return err
			}
		}
	}
	for memberID, pool := range snap.MemberQuotaPools {
		if _, err := exec.Exec(ctx, `
			INSERT INTO member_quota_pools (member_id, personal_quota)
			VALUES ($1, $2) ON CONFLICT (member_id) DO NOTHING
		`, memberID, pool.PersonalQuota); err != nil {
			return err
		}
	}
	return nil
}

func insertModels(ctx context.Context, exec tableWriter, models []types.ModelInfo) error {
	for _, model := range models {
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (
				id, provider, name, display_name, input_price, output_price, max_context, enabled
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, model.ID, model.Provider, model.Name, model.DisplayName,
			model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled); err != nil {
			return err
		}
		for _, cap := range model.Capabilities {
			if _, err := exec.Exec(ctx, `
				INSERT INTO model_capabilities (model_id, capability) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, model.ID, cap); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertRoutingRules(ctx context.Context, exec tableWriter, rules []types.RoutingRule) error {
	for _, rule := range rules {
		if _, err := exec.Exec(ctx, `
			INSERT INTO routing_rules (id, node_id, node_name, default_model, fallback_model, inherited)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING
		`, rule.ID, rule.NodeID, rule.NodeName, rule.DefaultModel, rule.FallbackModel, rule.Inherited); err != nil {
			return err
		}
		for _, modelName := range rule.AllowedModels {
			if _, err := exec.Exec(ctx, `
				INSERT INTO routing_rule_models (rule_id, model_name) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, rule.ID, modelName); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertKeys(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	for _, key := range snap.ProviderKeys {
		createdAt, err := timeparse.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var lastUsed *time.Time
		if key.LastUsed != nil {
			t, err := timeparse.Parse(*key.LastUsed)
			if err != nil {
				return err
			}
			lastUsed = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, relay_channel_id, status,
				balance, last_used, rotate_enabled, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, key.SecretKey, key.RelayChannelID,
			key.Status, key.Balance, lastUsed, key.RotateEnabled, createdAt); err != nil {
			return err
		}
	}
	for _, key := range snap.PlatformKeys {
		createdAt, err := timeparse.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := timeparse.Parse(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO platform_keys (
				id, name, key_prefix, full_key, member_id, member_name, app_name,
				budget_group_id, budget_group_name, status, quota, used, created_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO NOTHING
		`, key.ID, key.Name, key.KeyPrefix, key.FullKey, key.MemberID, key.MemberName,
			key.AppName, key.BudgetGroupID, key.BudgetGroupName, key.Status,
			key.Quota, key.Used, createdAt, expiresAt); err != nil {
			return err
		}
		for _, modelName := range key.ModelWhitelist {
			if _, err := exec.Exec(ctx, `
				INSERT INTO platform_key_models (platform_key_id, model_name) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, key.ID, modelName); err != nil {
				return err
			}
		}
	}
	for _, approval := range snap.Approvals {
		createdAt, err := timeparse.Parse(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := timeparse.Parse(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO key_approvals (
				id, type, applicant, applicant_id, department, reason, requested_quota,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO NOTHING
		`, approval.ID, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedQuota, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return err
		}
		for _, modelName := range approval.RequestedModels {
			if _, err := exec.Exec(ctx, `
				INSERT INTO key_approval_models (approval_id, model_name) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, approval.ID, modelName); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertAudit(ctx context.Context, exec tableWriter, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO audit_settings (id, content_retention_enabled)
		VALUES (1, $1) ON CONFLICT (id) DO NOTHING
	`, snap.AuditSettings.ContentRetentionEnabled); err != nil {
		return err
	}
	for _, log := range snap.OperationLogs {
		createdAt, err := timeparse.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO operation_logs (id, action, operator, operator_id, target, detail, ip, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO NOTHING
		`, log.ID, log.Action, log.Operator, log.OperatorID, log.Target, log.Detail, log.IP, createdAt); err != nil {
			return err
		}
	}
	for _, log := range snap.CallLogs {
		createdAt, err := timeparse.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO call_logs (
				id, caller, caller_id, caller_type, model, provider,
				input_tokens, output_tokens, latency_ms, status, cost,
				input_preview, output_preview, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO NOTHING
		`, log.ID, log.Caller, log.CallerID, log.CallerType, log.Model, log.Provider,
			log.InputTokens, log.OutputTokens, log.LatencyMs, log.Status, log.Cost,
			log.InputPreview, log.OutputPreview, createdAt); err != nil {
			return err
		}
	}
	return nil
}
