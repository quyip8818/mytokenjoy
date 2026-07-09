package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgBudgetRepo struct {
	db dbQuerier
}

func (r *pgBudgetRepo) AcquireBudgetLock(ctx context.Context) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", companyID)
	return err
}

func (r *pgBudgetRepo) GetGroupBudget(ctx context.Context, groupID string) (float64, float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var budget float64
	err := r.db.QueryRow(ctx, `
		SELECT budget FROM budget_groups WHERE company_id = $1 AND id = $2
	`, companyID, groupID).Scan(&budget)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return budget, 0, true, nil
}

func (r *pgBudgetRepo) Groups(ctx context.Context) ([]types.BudgetGroup, error) {
	companyID := store.CompanyID(ctx)

	// Query 1: All groups
	rows, err := r.db.Query(ctx, `SELECT id, name, budget FROM budget_groups WHERE company_id = $1 ORDER BY id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]types.BudgetGroup, 0)
	groupIndex := make(map[string]int) // id -> index in groups slice
	for rows.Next() {
		var g types.BudgetGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Budget); err != nil {
			return nil, err
		}
		g.Consumed = 0
		g.MemberIDs = []string{}
		g.DepartmentIDs = []string{}
		groupIndex[g.ID] = len(groups)
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}

	// Query 2: All members for all groups in this company
	memberRows, err := r.db.Query(ctx, `SELECT group_id, member_id FROM budget_group_members WHERE company_id = $1 ORDER BY group_id, member_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()
	for memberRows.Next() {
		var groupID, memberID string
		if err := memberRows.Scan(&groupID, &memberID); err != nil {
			return nil, err
		}
		if idx, ok := groupIndex[groupID]; ok {
			groups[idx].MemberIDs = append(groups[idx].MemberIDs, memberID)
		}
	}
	if err := memberRows.Err(); err != nil {
		return nil, err
	}

	// Query 3: All departments for all groups in this company
	deptRows, err := r.db.Query(ctx, `SELECT group_id, department_id FROM budget_group_departments WHERE company_id = $1 ORDER BY group_id, department_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var groupID, deptID string
		if err := deptRows.Scan(&groupID, &deptID); err != nil {
			return nil, err
		}
		if idx, ok := groupIndex[groupID]; ok {
			groups[idx].DepartmentIDs = append(groups[idx].DepartmentIDs, deptID)
		}
	}
	if err := deptRows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *pgBudgetRepo) SetGroups(ctx context.Context, groups []types.BudgetGroup) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneBudgetGroups(groups)
	ids := make([]string, len(cloned))
	for i, group := range cloned {
		ids[i] = group.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_groups (id, company_id, name, budget, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				updated_at = NOW()
		`, group.ID, companyID, group.Name, group.Budget); err != nil {
			return fmt.Errorf("upsert budget group %s: %w", group.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_members WHERE company_id = $1 AND group_id = $2`, companyID, group.ID); err != nil {
			return err
		}
		for _, memberID := range group.MemberIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO budget_group_members (company_id, group_id, member_id) VALUES ($1, $2, $3)
			`, companyID, group.ID, memberID); err != nil {
				return err
			}
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_departments WHERE company_id = $1 AND group_id = $2`, companyID, group.ID); err != nil {
			return err
		}
		for _, deptID := range group.DepartmentIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO budget_group_departments (company_id, group_id, department_id) VALUES ($1, $2, $3)
			`, companyID, group.ID, deptID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_members WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM budget_group_departments WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM budget_groups WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "budget_group_members", "group_id", companyID, ids); err != nil {
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "budget_group_departments", "group_id", companyID, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE platform_keys SET budget_group_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND budget_group_id IS NOT NULL AND NOT (budget_group_id = ANY($2))
	`, companyID, ids); err != nil {
		return fmt.Errorf("detach platform keys from pruned budget groups: %w", err)
	}
	return pruneByIDForCompany(ctx, r.db, "budget_groups", companyID, ids)
}

func (r *pgBudgetRepo) OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	companyID := store.CompanyID(ctx)
	var policy types.OverrunPolicyConfig
	err := r.db.QueryRow(ctx, `
		SELECT thresholds, notify_email, notify_phone, notify_im, block_message
		FROM overrun_policy WHERE company_id = $1
	`, companyID).Scan(&policy.Thresholds, &policy.NotifyEmail, &policy.NotifyPhone, &policy.NotifyIm, &policy.BlockMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.OverrunPolicyConfig{}, nil
		}
		return types.OverrunPolicyConfig{}, err
	}
	return policy, nil
}

func (r *pgBudgetRepo) SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO overrun_policy (company_id, thresholds, notify_email, notify_phone, notify_im, block_message, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			thresholds = EXCLUDED.thresholds,
			notify_email = EXCLUDED.notify_email,
			notify_phone = EXCLUDED.notify_phone,
			notify_im = EXCLUDED.notify_im,
			block_message = EXCLUDED.block_message,
			updated_at = NOW()
	`, companyID, policy.Thresholds, policy.NotifyEmail, policy.NotifyPhone, policy.NotifyIm, policy.BlockMessage)
	if err != nil {
		return fmt.Errorf("upsert overrun policy: %w", err)
	}
	return nil
}

func (r *pgBudgetRepo) AlertRules(ctx context.Context) ([]types.AlertRule, error) {
	companyID := store.CompanyID(ctx)

	// Query 1: All rules
	rows, err := r.db.Query(ctx, `
		SELECT ar.id, ar.node_id, n.name, ar.thresholds, ar.enabled
		FROM alert_rules ar
		JOIN org_nodes n ON n.company_id = ar.company_id AND n.id = ar.node_id
		WHERE ar.company_id = $1
		ORDER BY ar.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := make([]types.AlertRule, 0)
	ruleIndex := make(map[string]int) // id -> index in rules slice
	for rows.Next() {
		var rule types.AlertRule
		if err := rows.Scan(&rule.ID, &rule.NodeID, &rule.NodeName, &rule.Thresholds, &rule.Enabled); err != nil {
			return nil, err
		}
		rule.NotifyRoleIDs = []string{}
		ruleIndex[rule.ID] = len(rules)
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return rules, nil
	}

	// Query 2: All notify roles for all rules in this company
	roleRows, err := r.db.Query(ctx, `SELECT rule_id, role_id FROM alert_rule_notify_roles WHERE company_id = $1 ORDER BY rule_id, role_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var ruleID, roleID string
		if err := roleRows.Scan(&ruleID, &roleID); err != nil {
			return nil, err
		}
		if idx, ok := ruleIndex[ruleID]; ok {
			rules[idx].NotifyRoleIDs = append(rules[idx].NotifyRoleIDs, roleID)
		}
	}
	if err := roleRows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *pgBudgetRepo) SetAlertRules(ctx context.Context, rules []types.AlertRule) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneAlertRules(rules)
	ids := make([]string, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO alert_rules (id, company_id, node_id, thresholds, enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				thresholds = EXCLUDED.thresholds,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, rule.ID, companyID, rule.NodeID, rule.Thresholds, rule.Enabled); err != nil {
			return fmt.Errorf("upsert alert rule %s: %w", rule.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM alert_rule_notify_roles WHERE company_id = $1 AND rule_id = $2`, companyID, rule.ID); err != nil {
			return err
		}
		for _, roleID := range rule.NotifyRoleIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO alert_rule_notify_roles (company_id, rule_id, role_id) VALUES ($1, $2, $3)
			`, companyID, rule.ID, roleID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM alert_rule_notify_roles WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM alert_rules WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "alert_rule_notify_roles", "rule_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "alert_rules", companyID, ids)
}

func (r *pgBudgetRepo) BudgetApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT ba.id, ba.applicant_id, ba.applicant_name, ba.department_name, ba.amount, ba.reason,
			ba.status, ba.reject_reason, ba.created_at, ba.resolved_at,
			m.department_id
		FROM budget_approvals ba
		LEFT JOIN members m ON m.company_id = ba.company_id AND m.id = ba.applicant_id
		WHERE ba.company_id = $1 ORDER BY ba.created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.BudgetApproval, 0)
	for rows.Next() {
		var item types.BudgetApproval
		var applicantID *string
		var departmentID *string
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(
			&item.ID, &applicantID, &item.ApplicantName, &item.DepartmentName,
			&item.Amount, &item.Reason, &item.Status, &item.RejectReason,
			&createdAt, &resolvedAt, &departmentID,
		); err != nil {
			return nil, err
		}
		if applicantID != nil {
			item.ApplicantID = *applicantID
		}
		if departmentID != nil {
			item.DepartmentID = *departmentID
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		if resolvedAt != nil {
			s := formatSyncLogTime(*resolvedAt)
			item.ResolvedAt = &s
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgBudgetRepo) SetBudgetApprovals(ctx context.Context, items []types.BudgetApproval) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneBudgetApprovals(items)
	ids := make([]string, len(cloned))
	for i, approval := range cloned {
		ids[i] = approval.ID
		createdAt, err := parseAPITime(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := parseAPITime(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		var applicantID *string
		if approval.ApplicantID != "" {
			applicantID = &approval.ApplicantID
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_approvals (
				id, company_id, applicant_id, applicant_name, department_name,
				amount, reason, status, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (company_id, id) DO UPDATE SET
				applicant_id = EXCLUDED.applicant_id,
				applicant_name = EXCLUDED.applicant_name,
				department_name = EXCLUDED.department_name,
				amount = EXCLUDED.amount,
				reason = EXCLUDED.reason,
				status = EXCLUDED.status,
				reject_reason = EXCLUDED.reject_reason,
				created_at = EXCLUDED.created_at,
				resolved_at = EXCLUDED.resolved_at
		`, approval.ID, companyID, applicantID, approval.ApplicantName, approval.DepartmentName,
			approval.Amount, approval.Reason, approval.Status, approval.RejectReason,
			createdAt, resolvedAt); err != nil {
			return fmt.Errorf("upsert budget approval %s: %w", approval.ID, err)
		}
	}
	if len(ids) == 0 {
		_, err := r.db.Exec(ctx, `DELETE FROM budget_approvals WHERE company_id = $1`, companyID)
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "budget_approvals", companyID, ids)
}

func (r *pgBudgetRepo) UpdateBudgetApproval(ctx context.Context, id, status string, rejectReason *string, resolvedAt time.Time) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE budget_approvals
		SET status = $3, reject_reason = $4, resolved_at = $5
		WHERE company_id = $1 AND id = $2
	`, companyID, id, status, rejectReason, resolvedAt)
	return err
}
