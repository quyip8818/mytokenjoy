package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgBudgetRepo struct {
	db dbQuerier
}

func (r *pgBudgetRepo) Tree(ctx context.Context) ([]types.BudgetNode, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, parent_id, budget, consumed, reserved_pool, period, sort_order
		FROM budget_nodes WHERE company_id = $1 ORDER BY sort_order
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flat := make([]flatBudgetNode, 0)
	for rows.Next() {
		var row flatBudgetNode
		if err := rows.Scan(
			&row.ID, &row.Name, &row.ParentID, &row.Budget, &row.Consumed,
			&row.ReservedPool, &row.Period, &row.sortOrder,
		); err != nil {
			return nil, err
		}
		flat = append(flat, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneBudgetTree(buildBudgetTree(flat)), nil
}

func (r *pgBudgetRepo) SetTree(ctx context.Context, tree []types.BudgetNode) error {
	companyID := store.CompanyID(ctx)
	flat := flattenBudgetNodesWithOrder(store.CloneBudgetTree(tree))
	ids := make([]string, len(flat))
	for i, row := range flat {
		ids[i] = row.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_nodes (
				id, company_id, name, parent_id, budget, consumed, reserved_pool, period, sort_order, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				parent_id = EXCLUDED.parent_id,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				reserved_pool = EXCLUDED.reserved_pool,
				period = EXCLUDED.period,
				sort_order = EXCLUDED.sort_order,
				updated_at = NOW()
		`, row.ID, companyID, row.Name, row.ParentID, row.Budget, row.Consumed,
			row.ReservedPool, row.Period, row.sortOrder); err != nil {
			return fmt.Errorf("upsert budget node %s: %w", row.ID, err)
		}
	}
	return pruneByIDForCompany(ctx, r.db, "budget_nodes", companyID, ids)
}

func (r *pgBudgetRepo) Groups(ctx context.Context) ([]types.BudgetGroup, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, budget, consumed FROM budget_groups WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.BudgetGroup, 0)
	for rows.Next() {
		var group types.BudgetGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Budget, &group.Consumed); err != nil {
			return nil, err
		}
		memberRows, err := r.db.Query(ctx, `
			SELECT member_id FROM budget_group_members WHERE company_id = $1 AND group_id = $2 ORDER BY member_id
		`, companyID, group.ID)
		if err == nil {
			for memberRows.Next() {
				var memberID string
				if err := memberRows.Scan(&memberID); err == nil {
					group.MemberIDs = append(group.MemberIDs, memberID)
				}
			}
			memberRows.Close()
		}
		deptRows, err := r.db.Query(ctx, `
			SELECT department_id FROM budget_group_departments WHERE company_id = $1 AND group_id = $2 ORDER BY department_id
		`, companyID, group.ID)
		if err == nil {
			for deptRows.Next() {
				var deptID string
				if err := deptRows.Scan(&deptID); err == nil {
					group.DepartmentIDs = append(group.DepartmentIDs, deptID)
				}
			}
			deptRows.Close()
		}
		items = append(items, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneBudgetGroups(items), nil
}

func (r *pgBudgetRepo) SetGroups(ctx context.Context, groups []types.BudgetGroup) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneBudgetGroups(groups)
	ids := make([]string, len(cloned))
	for i, group := range cloned {
		ids[i] = group.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO budget_groups (id, company_id, name, budget, consumed, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				consumed = EXCLUDED.consumed,
				updated_at = NOW()
		`, group.ID, companyID, group.Name, group.Budget, group.Consumed); err != nil {
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
	rows, err := r.db.Query(ctx, `
		SELECT id, node_id, node_name, thresholds, enabled FROM alert_rules WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.AlertRule, 0)
	for rows.Next() {
		var rule types.AlertRule
		if err := rows.Scan(&rule.ID, &rule.NodeID, &rule.NodeName, &rule.Thresholds, &rule.Enabled); err != nil {
			return nil, err
		}
		roleRows, err := r.db.Query(ctx, `
			SELECT role_id FROM alert_rule_notify_roles WHERE company_id = $1 AND rule_id = $2 ORDER BY role_id
		`, companyID, rule.ID)
		if err == nil {
			for roleRows.Next() {
				var roleID string
				if err := roleRows.Scan(&roleID); err == nil {
					rule.NotifyRoleIDs = append(rule.NotifyRoleIDs, roleID)
				}
			}
			roleRows.Close()
		}
		items = append(items, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneAlertRules(items), nil
}

func (r *pgBudgetRepo) SetAlertRules(ctx context.Context, rules []types.AlertRule) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneAlertRules(rules)
	ids := make([]string, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO alert_rules (id, company_id, node_id, node_name, thresholds, enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				node_name = EXCLUDED.node_name,
				thresholds = EXCLUDED.thresholds,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, rule.ID, companyID, rule.NodeID, rule.NodeName, rule.Thresholds, rule.Enabled); err != nil {
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
