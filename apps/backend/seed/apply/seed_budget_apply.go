package apply

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
)

func insertSeedBudget(ctx context.Context, exec TableWriter, tid uuid.UUID, snap store.Snapshot) error {
	for _, project := range snap.Projects {
		if _, err := exec.Exec(ctx, `
			INSERT INTO projects (id, company_id, name, budget, owner_department_id)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (company_id, id) DO NOTHING
		`, project.ID, tid, project.Name, project.Budget, project.OwnerDepartmentID); err != nil {
			return err
		}
		for _, memberID := range project.MemberIDs {
			memberBudget := 0.0
			if project.MemberBudgets != nil {
				memberBudget = project.MemberBudgets[memberID]
			}
			if _, err := exec.Exec(ctx, `
				INSERT INTO project_members (company_id, project_id, member_id, member_budget) VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING
			`, tid, project.ID, memberID, memberBudget); err != nil {
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
			INSERT INTO alert_rules (id, company_id, node_id, thresholds, enabled)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (company_id, id) DO NOTHING
		`, rule.ID, tid, rule.NodeID, rule.Thresholds, rule.Enabled); err != nil {
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

func insertSeedBudgetConsumed(ctx context.Context, exec TableWriter, tid uuid.UUID, snap store.Snapshot) error {
	if snap.SeedAt.IsZero() {
		return fmt.Errorf("seed budget consumed require Snapshot.SeedAt")
	}
	periodKey := pkgbudget.RootPeriodKey(snap.OrgNodes, snap.SeedAt.UTC())
	memberConsumed := make(map[uuid.UUID]float64)
	for _, key := range snap.PlatformKeys {
		if key.Scope == types.PlatformKeyScopeMember && key.MemberID != nil && key.Consumed > 0 {
			memberConsumed[*key.MemberID] += key.Consumed
		}
	}
	for memberID, consumed := range memberConsumed {
		if err := insertBudgetConsumedRow(ctx, exec, tid, store.AxisKindMember, memberID, periodKey, consumed); err != nil {
			return fmt.Errorf("seed budget consumed member %s: %w", memberID, err)
		}
	}
	for _, project := range snap.Projects {
		if project.Consumed <= 0 {
			continue
		}
		if err := insertBudgetConsumedRow(ctx, exec, tid, store.AxisKindProject, project.ID, periodKey, project.Consumed); err != nil {
			return fmt.Errorf("seed budget consumed project %s: %w", project.ID, err)
		}
	}
	for _, key := range snap.PlatformKeys {
		if key.Consumed <= 0 {
			continue
		}
		if err := insertBudgetConsumedRow(ctx, exec, tid, store.AxisKindPlatformKey, key.ID, periodKey, key.Consumed); err != nil {
			return fmt.Errorf("seed budget consumed platform key %s: %w", key.ID, err)
		}
	}
	return nil
}

func insertBudgetConsumedRow(ctx context.Context, exec TableWriter, tid uuid.UUID, axisKind string, axisID uuid.UUID, periodKey string, consumed float64) error {
	_, err := exec.Exec(ctx, `
		INSERT INTO budget_consumed (company_id, axis_kind, axis_id, period_key, consumed, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (company_id, axis_kind, axis_id, period_key) DO NOTHING
	`, tid, axisKind, axisID, periodKey, consumed)
	return err
}

func insertSeedBudgetApprovals(ctx context.Context, exec TableWriter, tid uuid.UUID, approvals []types.BudgetApproval) error {
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
		var applicantID *uuid.UUID
		if approval.ApplicantID != uuid.Nil {
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
