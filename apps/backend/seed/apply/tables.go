package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

type TableWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func ApplyTables(ctx context.Context, exec TableWriter, snap store.Snapshot) error {
	if err := insertSeedCompany(ctx, exec, snap); err != nil {
		return err
	}
	tid := snap.Company.ID
	if err := insertSeedPermissions(ctx, exec, snap.Permissions); err != nil {
		return err
	}
	if err := insertSeedRoles(ctx, exec, tid, snap.Roles); err != nil {
		return err
	}
	roleIDByName := buildSeedRoleNameIndex(snap.Roles)
	if err := insertSeedOrgNodes(ctx, exec, tid, snap.OrgNodes); err != nil {
		return err
	}
	if err := insertSeedMembers(ctx, exec, tid, snap.Members, roleIDByName); err != nil {
		return err
	}
	if err := insertSeedOrgIntegration(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedBudget(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedModels(ctx, exec, tid, snap.Models); err != nil {
		return err
	}
	if err := insertSeedModelAllowlist(ctx, exec, tid, snap.ModelAllowlist); err != nil {
		return err
	}
	if err := insertSeedKeys(ctx, exec, tid, snap); err != nil {
		return err
	}
	if err := insertSeedAudit(ctx, exec, tid, snap); err != nil {
		return err
	}
	return nil
}

func insertSeedCompany(ctx context.Context, exec TableWriter, snap store.Snapshot) error {
	t := snap.Company
	if _, err := exec.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, t.ID, t.Slug, t.Name, t.Status); err != nil {
		return fmt.Errorf("seed company: %w", err)
	}
	return nil
}

func insertSeedPermissions(ctx context.Context, exec TableWriter, permissions []types.Permission) error {
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

func insertSeedRoles(ctx context.Context, exec TableWriter, tid int64, roles []types.Role) error {
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

func buildSeedRoleNameIndex(roles []types.Role) map[string]string {
	index := make(map[string]string, len(roles))
	for _, role := range roles {
		index[role.Name] = role.ID
	}
	return index
}

func insertSeedOrgNodes(ctx context.Context, exec TableWriter, tid int64, nodes []types.OrgNode) error {
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

func insertSeedModelAllowlist(ctx context.Context, exec TableWriter, tid int64, rows []store.ModelAllowlistRow) error {
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

func insertSeedMembers(ctx context.Context, exec TableWriter, tid int64, members []types.Member, roleIDByName map[string]string) error {
	demoHash := contract.DemoPasswordHash()
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

func insertSeedOrgIntegration(ctx context.Context, exec TableWriter, tid int64, snap store.Snapshot) error {
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

func insertSeedBudget(ctx context.Context, exec TableWriter, tid int64, snap store.Snapshot) error {
	for _, group := range snap.BudgetGroups {
		if _, err := exec.Exec(ctx, `
			INSERT INTO budget_groups (id, company_id, name, budget, consumed)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (company_id, id) DO NOTHING
		`, group.ID, tid, group.Name, group.Budget, group.Consumed); err != nil {
			return err
		}
		for _, memberID := range group.MemberIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO budget_group_members (company_id, group_id, member_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, group.ID, memberID); err != nil {
				return err
			}
		}
		for _, deptID := range group.DepartmentIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO budget_group_departments (company_id, group_id, department_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, group.ID, deptID); err != nil {
				return err
			}
		}
	}
	policy := snap.OverrunPolicy
	if _, err := exec.Exec(ctx, `
		INSERT INTO overrun_policy (company_id, thresholds, notify_email, notify_phone, notify_im, block_message)
		VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (company_id) DO NOTHING
	`, tid, policy.Thresholds, policy.NotifyEmail, policy.NotifyPhone, policy.NotifyIm, policy.BlockMessage); err != nil {
		return err
	}
	for _, rule := range snap.AlertRules {
		if _, err := exec.Exec(ctx, `
			INSERT INTO alert_rules (id, company_id, node_id, node_name, thresholds, enabled)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (company_id, id) DO NOTHING
		`, rule.ID, tid, rule.NodeID, rule.NodeName, rule.Thresholds, rule.Enabled); err != nil {
			return err
		}
		for _, roleID := range rule.NotifyRoleIDs {
			if _, err := exec.Exec(ctx, `
				INSERT INTO alert_rule_notify_roles (company_id, rule_id, role_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, rule.ID, roleID); err != nil {
				return err
			}
		}
	}
	return insertSeedBudgetApprovals(ctx, exec, tid, snap.BudgetApprovals)
}

func insertSeedBudgetApprovals(ctx context.Context, exec TableWriter, tid int64, approvals []types.BudgetApproval) error {
	for _, approval := range approvals {
		createdAt, err := pkgtime.Parse(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, parseErr := pkgtime.Parse(*approval.ResolvedAt)
			if parseErr != nil {
				return parseErr
			}
			resolvedAt = &t
		}
		var applicantID *string
		if approval.ApplicantID != "" {
			applicantID = &approval.ApplicantID
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO budget_approvals (
				id, company_id, applicant_id, applicant_name, department_name,
				amount, reason, status, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (company_id, id) DO NOTHING
		`, approval.ID, tid, applicantID, approval.ApplicantName, approval.DepartmentName,
			approval.Amount, approval.Reason, approval.Status, approval.RejectReason,
			createdAt, resolvedAt); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedKeys(ctx context.Context, exec TableWriter, tid int64, snap store.Snapshot) error {
	for _, key := range snap.ProviderKeys {
		createdAt, err := pkgtime.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var lastUsed *time.Time
		if key.LastUsed != nil {
			t, err := pkgtime.Parse(*key.LastUsed)
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
		createdAt, err := pkgtime.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := pkgtime.Parse(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO platform_keys (
				id, company_id, name, key_prefix, full_key, member_id,
				budget_group_id, status, quota, used, created_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (company_id, id) DO NOTHING
		`, key.ID, tid, key.Name, key.KeyPrefix, key.FullKey, key.MemberID,
			key.BudgetGroupID, key.Status,
			key.Quota, key.Used, createdAt, expiresAt); err != nil {
			return err
		}
	}
	for _, approval := range snap.Approvals {
		createdAt, err := pkgtime.Parse(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := pkgtime.Parse(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO key_approvals (
				id, company_id, type, applicant, applicant_id, department, reason, requested_quota,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO NOTHING
		`, approval.ID, tid, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedQuota, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedModels(ctx context.Context, exec TableWriter, tid int64, models []types.ModelInfo) error {
	for _, model := range models {
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (
				id, company_id, provider, name, display_name, model_type, description, visibility, endpoint,
				input_price, output_price, max_context, enabled
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO NOTHING
		`, model.ID, tid, model.Provider, model.Name, model.DisplayName,
			model.Type, model.Description, model.Visibility, model.Endpoint,
			model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled); err != nil {
			return err
		}
		for _, cap := range model.Capabilities {
			if _, err := exec.Exec(ctx, `
				INSERT INTO model_capabilities (company_id, model_id, capability) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, model.ID, cap); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertSeedAudit(ctx context.Context, exec TableWriter, tid int64, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO audit_settings (company_id, content_retention_enabled)
		VALUES ($1, $2) ON CONFLICT (company_id) DO NOTHING
	`, tid, snap.AuditSettings.ContentRetentionEnabled); err != nil {
		return err
	}
	for _, log := range snap.OperationLogs {
		createdAt, err := pkgtime.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		actorType := log.ActorType
		if actorType == "" {
			actorType = store.ActorTypeMember
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO operation_logs (id, company_id, action, operator, operator_id, actor_type, target, detail, ip, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (company_id, id) DO NOTHING
		`, log.ID, tid, log.Action, log.Operator, log.OperatorID, actorType, log.Target, log.Detail, log.IP, createdAt); err != nil {
			return err
		}
	}
	for _, entry := range snap.UsageLedger {
		detailJSON, err := json.Marshal(entry.CallDetail)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO usage_ledger (
				id, company_id, event_type, idempotency_key, amount_cny,
				department_id, member_id, budget_group_id, platform_key_id,
				source, occurred_at, model, input_tokens, output_tokens,
				call_detail, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (company_id, id) DO NOTHING
		`, entry.ID, tid, entry.EventType, entry.IdempotencyKey, entry.AmountCNY,
			entry.DepartmentID, entry.MemberID, entry.BudgetGroupID, entry.PlatformKeyID,
			entry.Source, entry.OccurredAt.UTC(), entry.Model, entry.InputTokens, entry.OutputTokens,
			detailJSON, entry.CreatedAt.UTC()); err != nil {
			return err
		}
	}
	return nil
}
