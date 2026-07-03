package seed

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func insertBudget(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
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
	return nil
}
